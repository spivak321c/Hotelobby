package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"hotel_lobby/internal/models"
	"hotel_lobby/internal/sse"

	"github.com/google/uuid"
)

// ReservationService orchestrates reservation creation, lookup, and cancellation.
type ReservationService struct {
	reservationRepo ReservationRepository
	bookingRepo     BookingRepository
	paymentRepo     PaymentRepository
	roomRepo        RoomRepositoryService
	inventoryRepo   InventoryRepositoryService
	pricingRepo     PricingRepositoryService
	roomTypeRepo    RoomTypeRepositoryService
	otpStore        OTPStore
	emailService    *EmailService
	sseHub          *sse.Hub
	inventorySvc    *InventoryService
}

func NewReservationService(
	rRepo ReservationRepository,
	bRepo BookingRepository,
	pRepo PaymentRepository,
	roomRepo RoomRepositoryService,
	invRepo InventoryRepositoryService,
	prRepo PricingRepositoryService,
	rtRepo RoomTypeRepositoryService,
	otpStore OTPStore,
	emailSvc *EmailService,
	sseHub *sse.Hub,
	invSvc *InventoryService,
) *ReservationService {
	return &ReservationService{
		reservationRepo: rRepo,
		bookingRepo:     bRepo,
		paymentRepo:     pRepo,
		roomRepo:        roomRepo,
		inventoryRepo:   invRepo,
		pricingRepo:     prRepo,
		roomTypeRepo:    rtRepo,
		otpStore:        otpStore,
		emailService:    emailSvc,
		sseHub:          sseHub,
		inventorySvc:    invSvc,
	}
}

func generateReference() string {
	id := uuid.New().String()[:8]
	return "HB-" + id
}

// Create validates the booking inputs, prices each room, persists the
// reservation + bookings + pending payment, and sends a confirmation email.
func (s *ReservationService) Create(ctx context.Context, input CreateReservationInput) (*models.Reservation, error) {
	if len(input.Bookings) == 0 {
		return nil, errors.New("at least one booking required")
	}
	if len(input.Bookings) > 4 {
		return nil, ErrMaxRoomsExceeded
	}

	var totalAmount float64
	var bookings []models.Booking

	for _, bi := range input.Bookings {
		if !bi.CheckOut.After(bi.CheckIn) {
			return nil, ErrInvalidBookingDates
		}

		room, err := s.roomRepo.FindByID(ctx, bi.RoomID)
		if err != nil {
			return nil, ErrRoomNotFound
		}
		if room.Status != "active" {
			return nil, ErrRoomNotAvailable
		}

		avail, err := s.roomRepo.IsAvailable(ctx, bi.RoomID, bi.CheckIn, bi.CheckOut)
		if err != nil {
			return nil, err
		}
		if !avail {
			return nil, ErrRoomNotAvailable
		}

		rate, err := s.getEffectiveRate(ctx, room.RoomTypeID, bi.CheckIn, bi.CheckOut, bi.BookingType)
		if err != nil {
			return nil, err
		}

		amount := calcBookingAmount(rate, bi.CheckIn, bi.CheckOut, bi.BookingType)
		totalAmount += amount

		bookings = append(bookings, models.Booking{
			ID:          uuid.New(),
			RoomID:      bi.RoomID,
			RoomTypeID:  room.RoomTypeID,
			StartsAt:    bi.CheckIn,
			EndsAt:      bi.CheckOut,
			BookingType: bi.BookingType,
			Amount:      amount,
			Status:      "pending",
		})
	}

	if input.IdempotencyKey != nil && *input.IdempotencyKey != "" {
		existing, err := s.reservationRepo.FindByIdempotencyKey(ctx, *input.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	reservation := &models.Reservation{
		ID:          uuid.New(),
		ReferenceCode: generateReference(),
		CustomerID:  input.CustomerID,
		GuestName:   input.GuestName,
		GuestEmail:  input.GuestEmail,
		GuestPhone:  input.GuestPhone,
		TotalAmount: totalAmount,
		Status:      "pending",
		IdempotencyKey: input.IdempotencyKey,
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, err
	}

	for i := range bookings {
		bookings[i].ReservationID = reservation.ID
		if err := s.bookingRepo.Create(ctx, &bookings[i]); err != nil {
			return nil, err
		}
		if input.SessionID != "" && s.inventorySvc != nil {
			s.inventorySvc.ReleaseHold(ctx, input.SessionID, bookings[i].RoomID.String())
		}
	}

	payment := &models.Payment{
		ID:                uuid.New(),
		ReservationID:     reservation.ID,
		Amount:            totalAmount,
		Provider:          input.PaymentMethod,
		ProviderReference: fmt.Sprintf("pending_%s_%s", reservation.ReferenceCode, uuid.New().String()[:8]),
		Status:            "pending",
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	for _, b := range bookings {
		for d := b.StartsAt; d.Before(b.EndsAt); d = d.AddDate(0, 0, 1) {
			s.inventoryRepo.IncrementBooked(ctx, b.RoomTypeID, d)
		}
		s.publishAvailabilityForBooking(ctx, b.RoomTypeID, b.StartsAt, b.EndsAt)
	}

	if s.sseHub != nil && reservation.CustomerID != nil {
		s.sseHub.PublishBookingUpdated(reservation.CustomerID.String(), reservation.ReferenceCode, "pending")
	}

	if input.SessionID != "" && s.emailService != nil {
		s.emailService.SendAbandonedBooking(reservation.GuestEmail, reservation.ReferenceCode)
	}

	return reservation, nil
}

// Lookup retrieves a reservation by reference code, verifying the email matches.
func (s *ReservationService) Lookup(ctx context.Context, reference, email string) (*models.Reservation, error) {
	reservation, err := s.reservationRepo.FindByReferenceCode(ctx, reference)
	if err != nil {
		return nil, ErrReservationNotFound
	}
	if reservation.GuestEmail != email {
		return nil, ErrReservationNotFound
	}
	return reservation, nil
}

// Cancel applies an OTP-verified cancellation to a reservation.
func (s *ReservationService) Cancel(ctx context.Context, reference, otp, reason string) error {
	key := fmt.Sprintf("cancel_otp:%s", reference)
	storedOTP, err := s.otpStore.Get(ctx, key)
	if err != nil || storedOTP != otp {
		return ErrInvalidOTP
	}
	defer s.otpStore.Del(ctx, key)

	reservation, err := s.reservationRepo.FindByReferenceCode(ctx, reference)
	if err != nil {
		return ErrReservationNotFound
	}
	if reservation.Status == "cancelled" {
		return ErrAlreadyCancelled
	}

	reservation.Status = "cancelled"
	reservation.CancellationReason = reason
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return err
	}

	bookings, err := s.bookingRepo.FindByReservationID(ctx, reservation.ID)
	if err == nil {
		for _, b := range bookings {
			for d := b.StartsAt; d.Before(b.EndsAt); d = d.AddDate(0, 0, 1) {
				s.inventoryRepo.DecrementBooked(ctx, b.RoomTypeID, d)
			}
			s.publishAvailabilityForBooking(ctx, b.RoomTypeID, b.StartsAt, b.EndsAt)
		}
	}

	if s.sseHub != nil && reservation.CustomerID != nil {
		s.sseHub.PublishBookingUpdated(reservation.CustomerID.String(), reference, "cancelled")
	}

	if s.emailService != nil {
		s.emailService.SendCancellationConfirmation(reservation.GuestEmail, reference)
	}
	return nil
}

func (s *ReservationService) FindAll(ctx context.Context, status string, from, to time.Time) ([]models.Reservation, error) {
	return s.reservationRepo.FindAll(ctx, status, from, to)
}

func (s *ReservationService) FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error) {
	return s.reservationRepo.FindByID(ctx, id)
}

var validTransitions = map[string]map[string]bool{
	"pending":    {"confirmed": true, "cancelled": true},
	"confirmed":  {"checked_in": true, "cancelled": true, "refunded": true},
	"checked_in": {"checked_out": true, "cancelled": true},
	"checked_out": {},
	"cancelled":  {"refunded": true},
	"refunded":   {},
}

func (s *ReservationService) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reason string) error {
	res, err := s.reservationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !validTransitions[res.Status][status] {
		return fmt.Errorf("cannot transition from %s to %s", res.Status, status)
	}

	// When cancelling, unlock inventory, cancel bookings, and push real-time events.
	if status == "cancelled" && res.Status != "cancelled" {
		res.CancellationReason = reason
		if bookings, err := s.bookingRepo.FindByReservationID(ctx, id); err == nil {
			for i := range bookings {
				b := &bookings[i]
				b.Status = "cancelled"
				s.bookingRepo.Update(ctx, b)

				for d := b.StartsAt; d.Before(b.EndsAt); d = d.AddDate(0, 0, 1) {
					s.inventoryRepo.DecrementBooked(ctx, b.RoomTypeID, d)
				}
				s.publishAvailabilityForBooking(ctx, b.RoomTypeID, b.StartsAt, b.EndsAt)
			}
		}

		if s.sseHub != nil && res.CustomerID != nil {
			s.sseHub.PublishBookingUpdated(res.CustomerID.String(), res.ReferenceCode, "cancelled")
		}

		if s.emailService != nil {
			s.emailService.SendCancellationConfirmation(res.GuestEmail, res.ReferenceCode)
		}
	}

	res.Status = status
	return s.reservationRepo.Update(ctx, res)
}

func (s *ReservationService) FindAllBookingsByReservation(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	return s.bookingRepo.FindByReservationID(ctx, reservationID)
}

func (s *ReservationService) publishAvailabilityForBooking(ctx context.Context, roomTypeID uuid.UUID, checkIn, checkOut time.Time) {
	if s.sseHub == nil {
		return
	}
	rooms, err := s.roomRepo.FindAll(ctx, &roomTypeID, "")
	if err != nil {
		log.Printf("publishAvailabilityForBooking: FindAll(roomTypeID=%s): %v", roomTypeID, err)
		return
	}
	for d := checkIn; d.Before(checkOut); d = d.AddDate(0, 0, 1) {
		dayEnd := d.Add(24 * time.Hour)
		roomAvail := make([]sse.RoomAvailability, 0, len(rooms))
		for _, room := range rooms {
			ok, _ := s.roomRepo.IsAvailable(ctx, room.ID, d, dayEnd)
			roomAvail = append(roomAvail, sse.RoomAvailability{
				RoomID:     room.ID.String(),
				RoomNumber: room.RoomNumber,
				Available:  ok && room.Status == "active",
			})
		}
		s.sseHub.PublishAvailability(roomTypeID.String(), d.Format("2006-01-02"), roomAvail)
	}
}
