package services

import (
	"context"
	"errors"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

// Sentinel errors for the reservation domain.
var (
	ErrReservationNotFound = errors.New("reservation not found")
	ErrMaxRoomsExceeded    = errors.New("max 4 rooms per reservation")
	ErrRoomNotAvailable    = errors.New("room not available for selected dates")
	ErrInvalidBookingDates = errors.New("check-out must be after check-in")
	ErrInvalidOTP          = errors.New("invalid or expired OTP")
	ErrAlreadyCancelled    = errors.New("reservation already cancelled")
)

const otpTTL = 15 * time.Minute

// OTPStore persists short-lived OTP codes.
type OTPStore interface {
	Set(ctx context.Context, key, otp string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

// ReservationRepository is the persistence interface for Reservation records.
type ReservationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	FindByReferenceCode(ctx context.Context, reference string) (*models.Reservation, error)
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error)
	FindAll(ctx context.Context, status string, from, to time.Time) ([]models.Reservation, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*models.Reservation, error)
	Create(ctx context.Context, r *models.Reservation) error
	Update(ctx context.Context, r *models.Reservation) error
}

// BookingRepository is the persistence interface for Booking records.
type BookingRepository interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error)
	Create(ctx context.Context, b *models.Booking) error
	Update(ctx context.Context, b *models.Booking) error
}

// PaymentRepository is the persistence interface for Payment records
// as seen by the reservation service.
type PaymentRepository interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) (*models.Payment, error)
	Create(ctx context.Context, p *models.Payment) error
	Update(ctx context.Context, p *models.Payment) error
}

// RoomRepositoryService is the minimal room-lookup interface needed by the
// reservation service (avoids circular dependency on the full RoomService).
type RoomRepositoryService interface {
	FindAll(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Room, error)
	IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error)
}

// InventoryRepositoryService is the minimal inventory interface needed.
type InventoryRepositoryService interface {
	FindByRoomTypeAndDate(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error)
	FindByRoomTypeAndDateRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error)
	IncrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
	DecrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
}

// PricingRepositoryService is the minimal pricing interface needed.
type PricingRepositoryService interface {
	FindByRoomTypeID(ctx context.Context, roomTypeID uuid.UUID) ([]models.RoomPricing, error)
}

// RoomTypeRepositoryService is the minimal room-type interface needed.
type RoomTypeRepositoryService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.RoomType, error)
}

// BookingInput carries the per-room details from the HTTP layer.
type BookingInput struct {
	RoomID            uuid.UUID `json:"room_id"`
	CheckIn           time.Time `json:"check_in"`
	CheckOut          time.Time `json:"check_out"`
	BookingType       string    `json:"booking_type"`
	ExpectedOccupants int       `json:"expected_occupants"`
}

// CreateReservationInput is the validated payload for creating a reservation.
type CreateReservationInput struct {
	GuestName      string         `json:"guest_name"`
	GuestEmail     string         `json:"guest_email"`
	GuestPhone     string         `json:"guest_phone"`
	CustomerID     *uuid.UUID     `json:"customer_id,omitempty"`
	SessionID      string         `json:"session_id,omitempty"`
	Bookings       []BookingInput `json:"bookings"`
	PaymentMethod  string         `json:"payment_method"`
	IdempotencyKey *string        `json:"idempotency_key,omitempty"`
}
