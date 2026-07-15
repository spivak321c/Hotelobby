package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "staging.crossmint.com" {
		body := `{"orderId": "cm_mock_123", "phase": "processing"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}
	if req.URL.Host == "api.paystack.co" {
		body := `{"status":true,"data":{"reference":"ps_mock_123"}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}
	// Fallback for other requests
	return http.DefaultTransport.RoundTrip(req)
}

// mockPaymentUpdater satisfies the PaymentUpdater interface.
// It is distinct from the mockPaymentRepo in reservation_service_test.go,
// which satisfies the narrower PaymentRepository interface used by ReservationService.
type mockPaymentUpdater struct {
	payments          map[uuid.UUID]*models.Payment
	byProviderRef     map[string]*models.Payment

	// capture the last UpdateStatus call for assertion
	lastUpdateID     uuid.UUID
	lastUpdateStatus string
}

func (m *mockPaymentUpdater) FindByReservationID(ctx context.Context, reservationID uuid.UUID) (*models.Payment, error) {
	for _, p := range m.payments {
		if p.ReservationID == reservationID {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockPaymentUpdater) FindByProviderReference(ctx context.Context, txRef string) (*models.Payment, error) {
	p, ok := m.byProviderRef[txRef]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockPaymentUpdater) Create(ctx context.Context, p *models.Payment) error {
	if m.payments == nil {
		m.payments = make(map[uuid.UUID]*models.Payment)
	}
	if m.byProviderRef == nil {
		m.byProviderRef = make(map[string]*models.Payment)
	}
	m.payments[p.ID] = p
	if p.ProviderReference != "" {
		m.byProviderRef[p.ProviderReference] = p
	}
	return nil
}

func (m *mockPaymentUpdater) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.lastUpdateID = id
	m.lastUpdateStatus = status
	p, ok := m.payments[id]
	if !ok {
		return errors.New("not found")
	}
	p.Status = status
	return nil
}

// newTestPaymentService constructs a PaymentService wired to the given mock.
// Pass dummy strings for keys to exercise the mocked paths.
func newTestPaymentService(repo *mockPaymentUpdater) *PaymentService {
	http.DefaultTransport = &mockTransport{}
	return NewPaymentService("dummy_ps_secret", "dummy_ps_public", "dummy_ps_webhook", "dummy_cm_key", "dummy_cm_proj", "dummy_cm_webhook", repo, nil, nil, nil)
}

func TestPaymentService_ProcessPayment_card_noKey(t *testing.T) {
	repo := &mockPaymentUpdater{
		payments:      make(map[uuid.UUID]*models.Payment),
		byProviderRef: make(map[string]*models.Payment),
	}
	svc := newTestPaymentService(repo)

	result, err := svc.ProcessPayment(context.Background(), ProcessPaymentInput{
		ReservationID: uuid.New(),
		Method:        "card",
		Amount:        150.00,
		Email:         "guest@test.com",
	})
	if err != nil {
		t.Fatalf("ProcessPayment card noKey: %v", err)
	}
	if result.Status != "processing" {
		t.Errorf("expected processing, got %s", result.Status)
	}
	if result.ProviderReference == "" {
		t.Error("expected non-empty ProviderReference")
	}
}

func TestPaymentService_ProcessPayment_crypto_noKey(t *testing.T) {
	repo := &mockPaymentUpdater{
		payments:      make(map[uuid.UUID]*models.Payment),
		byProviderRef: make(map[string]*models.Payment),
	}
	svc := newTestPaymentService(repo)

	result, err := svc.ProcessPayment(context.Background(), ProcessPaymentInput{
		ReservationID: uuid.New(),
		Method:        "crypto",
		Amount:        200.00,
		Email:         "guest@test.com",
	})
	if err != nil {
		t.Fatalf("ProcessPayment crypto noKey: %v", err)
	}
	if result.Status != "processing" {
		t.Errorf("expected processing, got %s", result.Status)
	}
	if result.ProviderReference == "" {
		t.Error("expected non-empty ProviderReference")
	}
}

func TestPaymentService_ProcessPayment_invalidMethod(t *testing.T) {
	repo := &mockPaymentUpdater{
		payments:      make(map[uuid.UUID]*models.Payment),
		byProviderRef: make(map[string]*models.Payment),
	}
	svc := newTestPaymentService(repo)

	_, err := svc.ProcessPayment(context.Background(), ProcessPaymentInput{
		ReservationID: uuid.New(),
		Method:        "cash",
		Amount:        50.00,
	})
	if err != ErrInvalidPaymentMethod {
		t.Fatalf("expected ErrInvalidPaymentMethod, got %v", err)
	}
}

func TestPaymentService_CheckPayment_found(t *testing.T) {
	paymentID := uuid.New()
	reservationID := uuid.New()
	txRef := "PS_abc12345"

	seeded := &models.Payment{
		ID:                paymentID,
		ReservationID:     reservationID,
		Amount:            300.00,
		Provider:          "card",
		ProviderReference: txRef,
		Status:            "succeeded",
	}
	repo := &mockPaymentUpdater{
		payments:      map[uuid.UUID]*models.Payment{paymentID: seeded},
		byProviderRef: map[string]*models.Payment{txRef: seeded},
	}
	svc := newTestPaymentService(repo)

	p, err := svc.CheckPayment(context.Background(), txRef)
	if err != nil {
		t.Fatalf("CheckPayment: %v", err)
	}
	if p.ID != paymentID {
		t.Errorf("expected payment ID %s, got %s", paymentID, p.ID)
	}
	if p.Status != "succeeded" {
		t.Errorf("expected succeeded, got %s", p.Status)
	}
}

func TestPaymentService_CheckPayment_notFound(t *testing.T) {
	repo := &mockPaymentUpdater{
		payments:      make(map[uuid.UUID]*models.Payment),
		byProviderRef: make(map[string]*models.Payment),
	}
	svc := newTestPaymentService(repo)

	_, err := svc.CheckPayment(context.Background(), "no-such-ref")
	if err != ErrPaymentNotFound {
		t.Fatalf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestPaymentService_UpdatePaymentByTxRef(t *testing.T) {
	paymentID := uuid.New()
	txRef := "PS_update99"

	seeded := &models.Payment{
		ID:     paymentID,
		ProviderReference:  txRef,
		Status: "processing",
	}
	repo := &mockPaymentUpdater{
		payments:      map[uuid.UUID]*models.Payment{paymentID: seeded},
		byProviderRef: map[string]*models.Payment{txRef: seeded},
	}
	svc := newTestPaymentService(repo)

	if err := svc.UpdatePaymentByTxRef(context.Background(), txRef, "succeeded"); err != nil {
		t.Fatalf("UpdatePaymentByTxRef: %v", err)
	}
	if repo.lastUpdateID != paymentID {
		t.Errorf("expected UpdateStatus called with ID %s, got %s", paymentID, repo.lastUpdateID)
	}
	if repo.lastUpdateStatus != "succeeded" {
		t.Errorf("expected UpdateStatus called with 'succeeded', got %s", repo.lastUpdateStatus)
	}
}
