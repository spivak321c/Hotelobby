package repositories

import (
	"context"
	"errors"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

// Sentinels for constraint-to-domain-error translation (PRD DB Engineering Rule 13).
var (
	ErrRoomUnavailable    = errors.New("room unavailable for requested time range (exclusion violation)")
	ErrDuplicatePayment   = errors.New("duplicate payment (unique violation)")
	ErrConstraintViolation = errors.New("constraint violation")
)

type RoomTypeRepository interface {
	FindAll(ctx context.Context) ([]models.RoomType, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.RoomType, error)
	Create(ctx context.Context, rt *models.RoomType) error
	Update(ctx context.Context, rt *models.RoomType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoomRepository interface {
	FindAll(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Room, error)
	Create(ctx context.Context, r *models.Room) error
	Update(ctx context.Context, r *models.Room) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountActiveBookings(ctx context.Context, roomID uuid.UUID) (int, error)
	IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error)
	CountRooms(ctx context.Context) (int, error)
}

type RoomImageRepository interface {
	FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]models.RoomImage, error)
	Create(ctx context.Context, img *models.RoomImage) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetPrimary(ctx context.Context, id uuid.UUID) error
	Reorder(ctx context.Context, roomID uuid.UUID, ids []uuid.UUID) error
}

type RoomPricingRepository interface {
	FindAll(ctx context.Context, roomTypeID *uuid.UUID) ([]models.RoomPricing, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.RoomPricing, error)
	FindByRoomTypeID(ctx context.Context, roomTypeID uuid.UUID) ([]models.RoomPricing, error)
	Create(ctx context.Context, rp *models.RoomPricing) error
	Update(ctx context.Context, rp *models.RoomPricing) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoomTypeInventoryRepository interface {
	FindByRoomTypeAndDate(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error)
	FindByRoomTypeAndDateRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error)
	IncrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
	DecrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
	SetInventory(ctx context.Context, roomTypeID uuid.UUID, date time.Time, totalRooms, bookedRooms int) error
}

type CustomerRepository interface {
	FindAll(ctx context.Context) ([]models.Customer, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	FindByEmail(ctx context.Context, email string) (*models.Customer, error)
	Create(ctx context.Context, c *models.Customer) error
	Update(ctx context.Context, c *models.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdminRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	FindByEmail(ctx context.Context, email string) (*models.Admin, error)
	FindAll(ctx context.Context) ([]models.Admin, error)
	Create(ctx context.Context, a *models.Admin) error
	Update(ctx context.Context, a *models.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ReservationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	FindByReferenceCode(ctx context.Context, referenceCode string) (*models.Reservation, error)
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error)
	FindAll(ctx context.Context, status string, from, to time.Time) ([]models.Reservation, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*models.Reservation, error)
	Create(ctx context.Context, r *models.Reservation) error
	Update(ctx context.Context, r *models.Reservation) error
}

type BookingRepository interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error)
	FindByReservationIDBatch(ctx context.Context, reservationIDs []uuid.UUID) ([]models.Booking, error)
	Create(ctx context.Context, b *models.Booking) error
	Update(ctx context.Context, b *models.Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	FindAvailableRoom(ctx context.Context, roomTypeID uuid.UUID, startsAt, endsAt time.Time) (uuid.UUID, error)
}

type PaymentRepository interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) (*models.Payment, error)
	FindByProviderReference(ctx context.Context, providerRef string) (*models.Payment, error)
	Create(ctx context.Context, p *models.Payment) error
	Update(ctx context.Context, p *models.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}