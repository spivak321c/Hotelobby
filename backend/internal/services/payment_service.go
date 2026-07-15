package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"hotel_lobby/internal/models"
	"hotel_lobby/internal/providers/crossmint"
	"hotel_lobby/internal/providers/paystack"

	"github.com/google/uuid"
)

var (
	ErrPaymentFailed      = errors.New("payment failed")
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrInvalidPaymentMethod = errors.New("invalid payment method; use 'card' or 'crypto'")
)

type PaymentService struct {
	paystackClient      *paystack.Client
	crossmintClient     *crossmint.Client
	paymentRepo         PaymentUpdater
	reservationRepo     WebhookReservationRepo
	bookingRepo         WebhookBookingRepo
	emailService        *EmailService
	paystackWebhookSec  string
	crossmintWebhookSec string
}

type PaymentUpdater interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) (*models.Payment, error)
	FindByProviderReference(ctx context.Context, txRef string) (*models.Payment, error)
	Create(ctx context.Context, p *models.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type WebhookReservationRepo interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	Update(ctx context.Context, r *models.Reservation) error
}

type WebhookBookingRepo interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

func NewPaymentService(
	paystackSecret, paystackPublic, paystackWebhookSec, crossmintKey, crossmintProject, crossmintWebhookSec string,
	paymentRepo PaymentUpdater,
	reservationRepo WebhookReservationRepo,
	bookingRepo WebhookBookingRepo,
	emailService *EmailService,
) *PaymentService {
	return &PaymentService{
		paystackClient:      paystack.NewClient(paystackSecret, paystackPublic),
		crossmintClient:     crossmint.NewClient(crossmintKey, crossmintProject),
		paymentRepo:         paymentRepo,
		reservationRepo:     reservationRepo,
		bookingRepo:         bookingRepo,
		emailService:        emailService,
		paystackWebhookSec:  paystackWebhookSec,
		crossmintWebhookSec: crossmintWebhookSec,
	}
}

type ProcessPaymentInput struct {
	ReservationID uuid.UUID
	Method        string // "card" or "crypto"
	Amount        float64
	Email         string
}

type PaymentResult struct {
	Status             string
	ProviderReference  string
	Error              string
}

func (s *PaymentService) ProcessPayment(ctx context.Context, input ProcessPaymentInput) (*PaymentResult, error) {
	switch input.Method {
	case "card":
		return s.processCardPayment(ctx, input)
	case "crypto":
		return s.processCryptoPayment(ctx, input)
	default:
		return nil, ErrInvalidPaymentMethod
	}
}

func (s *PaymentService) processCardPayment(ctx context.Context, input ProcessPaymentInput) (*PaymentResult, error) {
	ref := fmt.Sprintf("PS_%s", uuid.New().String()[:8])
	amountKobo := int(input.Amount * 100)

	req := paystack.InitializeRequest{
		Email:     input.Email,
		Amount:    amountKobo,
		Reference: ref,
	}

	resp, err := s.paystackClient.InitializeTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("paystack initialization failed: %w", err)
	}

	return &PaymentResult{
		Status:            "processing",
		ProviderReference: resp.Data.Reference,
	}, nil
}

func (s *PaymentService) processCryptoPayment(ctx context.Context, input ProcessPaymentInput) (*PaymentResult, error) {
	req := crossmint.OrderRequest{
		Recipient: struct {
			Email string `json:"email"`
		}{Email: input.Email},
		Payment: struct {
			Method       string `json:"method"`
			Currency     string `json:"currency"`
			ReceiptEmail string `json:"receiptEmail,omitempty"`
		}{Method: "card", Currency: "usd", ReceiptEmail: input.Email},
		LineItems: []crossmint.LineItem{
			{
				CollectionLocator: "solana:dummy_collection",
				CallData: struct {
					TotalPrice string `json:"totalPrice"`
					Quantity   int    `json:"quantity"`
				}{TotalPrice: fmt.Sprintf("%.2f", input.Amount), Quantity: 1},
			},
		},
	}

	resp, err := s.crossmintClient.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("crossmint order creation failed: %w", err)
	}

	return &PaymentResult{
		Status:            "processing",
		ProviderReference: resp.OrderID,
	}, nil
}

func (s *PaymentService) VerifyPaystackSignature(body []byte, signature string) bool {
	if s.paystackWebhookSec == "" {
		return false
	}
	mac := hmac.New(sha512.New, []byte(s.paystackWebhookSec))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// HandlePaystackWebhook verifies the webhook signature, processes the event,
// and on charge.success updates payment + reservation status and sends
// the confirmation email.
func (s *PaymentService) HandlePaystackWebhook(ctx context.Context, body []byte, signature string) error {
	if !s.VerifyPaystackSignature(body, signature) {
		return ErrPaymentFailed
	}

	var payload struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("webhook parse failed: %w", err)
	}

	if payload.Event != "charge.success" {
		return nil // ignore non-success events
	}

	txRef := payload.Data.Reference

	p, err := s.paymentRepo.FindByProviderReference(ctx, txRef)
	if err != nil {
		return fmt.Errorf("find payment: %w", err)
	}

	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, "succeeded"); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	reservation, err := s.reservationRepo.FindByID(ctx, p.ReservationID)
	if err != nil {
		return fmt.Errorf("find reservation: %w", err)
	}

	reservation.Status = "confirmed"
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return fmt.Errorf("update reservation: %w", err)
	}

	bookings, err := s.bookingRepo.FindByReservationID(ctx, reservation.ID)
	if err != nil {
		return fmt.Errorf("find bookings: %w", err)
	}
	for _, b := range bookings {
		if err := s.bookingRepo.UpdateStatus(ctx, b.ID, "confirmed"); err != nil {
			return fmt.Errorf("update booking %s: %w", b.ID, err)
		}
	}

	if s.emailService != nil {
		roomNames := ""
		if len(bookings) > 0 {
			roomNames = bookings[0].ID.String()
		}
		s.emailService.SendConfirmation(
			reservation.GuestEmail, reservation.ReferenceCode, reservation.GuestName,
			"", "", roomNames,
		)
	}

	return nil
}

// HandleCrossmintWebhook verifies the Crossmint webhook signature and processes
// order status updates. On order:succeeded it marks payment + reservation as confirmed.
func (s *PaymentService) HandleCrossmintWebhook(ctx context.Context, body []byte, signature string) error {
	if s.crossmintWebhookSec == "" {
		return fmt.Errorf("crossmint webhook secret not configured")
	}

	mac := hmac.New(sha256.New, []byte(s.crossmintWebhookSec))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrPaymentFailed
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			OrderID string `json:"orderId"`
			Phase   string `json:"phase"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("crossmint webhook parse failed: %w", err)
	}

	if payload.Type != "order:succeeded" {
		return nil
	}

	p, err := s.paymentRepo.FindByProviderReference(ctx, payload.Data.OrderID)
	if err != nil {
		return fmt.Errorf("find payment by order id %s: %w", payload.Data.OrderID, err)
	}

	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, "succeeded"); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	reservation, err := s.reservationRepo.FindByID(ctx, p.ReservationID)
	if err != nil {
		return fmt.Errorf("find reservation: %w", err)
	}

	reservation.Status = "confirmed"
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return fmt.Errorf("update reservation: %w", err)
	}

	bookings, err := s.bookingRepo.FindByReservationID(ctx, reservation.ID)
	if err != nil {
		return fmt.Errorf("find bookings: %w", err)
	}
	for _, b := range bookings {
		if err := s.bookingRepo.UpdateStatus(ctx, b.ID, "confirmed"); err != nil {
			return fmt.Errorf("update booking %s: %w", b.ID, err)
		}
	}

	if s.emailService != nil {
		s.emailService.SendConfirmation(
			reservation.GuestEmail, reservation.ReferenceCode, reservation.GuestName,
			"", "", "",
		)
	}

	return nil
}

func (s *PaymentService) VerifyPaystackWebhook(body []byte, signature string) (txRef string, status string, err error) {
	var payload struct {
		Data struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
		} `json:"data"`
		Event string `json:"event"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", err
	}
	return payload.Data.Reference, payload.Data.Status, nil
}

func (s *PaymentService) CheckPayment(ctx context.Context, txRef string) (*models.Payment, error) {
	p, err := s.paymentRepo.FindByProviderReference(ctx, txRef)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return p, nil
}

func (s *PaymentService) UpdatePaymentByTxRef(ctx context.Context, txRef, status string) error {
	p, err := s.paymentRepo.FindByProviderReference(ctx, txRef)
	if err != nil {
		return ErrPaymentNotFound
	}
	return s.paymentRepo.UpdateStatus(ctx, p.ID, status)
}