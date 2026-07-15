package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomType struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	BaseRateDaily  float64   `json:"base_rate_daily"`
	BaseRateHourly float64   `json:"base_rate_hourly"`
	MaxOccupancy   int       `json:"max_occupancy"`
	IsFeatured     bool      `json:"is_featured"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Room struct {
	ID          uuid.UUID `json:"id"`
	RoomTypeID  uuid.UUID `json:"room_type_id"`
	RoomNumber  string    `json:"room_number"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomWithImages struct {
	Room   Room        `json:"room"`
	Images []RoomImage `json:"images"`
}

type RoomImage struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	URL       string    `json:"url"`
	IsPrimary bool      `json:"is_primary"`
	SortOrder int       `json:"sort_order"`
}

type RoomPricing struct {
	ID             uuid.UUID      `json:"id"`
	RoomTypeID     uuid.UUID      `json:"room_type_id"`
	RateType       string         `json:"rate_type"`
	Rate           float64        `json:"rate"`
	EffectiveRange Daterange      `json:"effective_range"`
}

type Daterange struct {
	Lower time.Time `json:"lower"`
	Upper time.Time `json:"upper"`
	Bounds string   `json:"bounds"`
}

type RoomTypeInventory struct {
	RoomTypeID uuid.UUID `json:"room_type_id"`
	Date       time.Time `json:"date"`
	TotalRooms int       `json:"total_rooms"`
	BookedRooms int       `json:"booked_rooms"`
}

type Customer struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Admin struct {
	ID           uuid.UUID `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Reservation struct {
	ID               uuid.UUID  `json:"id"`
	ReferenceCode    string     `json:"reference_code"`
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	GuestName        string     `json:"guest_name"`
	GuestEmail       string     `json:"guest_email"`
	GuestPhone       string     `json:"guest_phone"`
	TotalAmount      float64    `json:"total_amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	CancellationReason string   `json:"cancellation_reason,omitempty"`
	IdempotencyKey   *string    `json:"idempotency_key,omitempty"`
	CreatedByAdminID *uuid.UUID `json:"created_by_admin_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Booking struct {
	ID              uuid.UUID `json:"id"`
	ReservationID   uuid.UUID `json:"reservation_id"`
	RoomID          uuid.UUID `json:"room_id"`
	RoomTypeID      uuid.UUID `json:"room_type_id"`
	BookingType     string    `json:"booking_type"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Status          string    `json:"status"`
	Amount          float64   `json:"amount"`
}

type Payment struct {
	ID                uuid.UUID `json:"id"`
	ReservationID     uuid.UUID `json:"reservation_id"`
	Provider          string    `json:"provider"`
	ProviderReference string    `json:"provider_reference"`
	Status            string    `json:"status"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Metadata          []byte    `json:"metadata,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}