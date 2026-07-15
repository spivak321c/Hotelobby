package handlers

import "time"

type RegisterRequest struct {
	Name     string `json:"name"     validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RoomTypeResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	BaseRateHourly float64 `json:"base_rate_hourly"`
	BaseRateDaily  float64 `json:"base_rate_daily"`
	MaxOccupancy   int     `json:"max_occupancy"`
	IsFeatured     bool    `json:"is_featured"`
}

type RoomResponse struct {
	ID               string `json:"id"`
	RoomTypeID       string `json:"room_type_id"`
	RoomNumber       string `json:"room_number"`
	Status           string `json:"status"`
	UpcomingBookings int    `json:"upcoming_bookings"`
}

type RoomImageResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

type RoomDetailResponse struct {
	Room           RoomResponse        `json:"room"`
	RoomTypeName   string              `json:"room_type_name"`
	BaseRateDaily  float64             `json:"base_rate_daily"`
	BaseRateHourly float64             `json:"base_rate_hourly"`
	Images         []RoomImageResponse `json:"images"`
}

type PricingResponse struct {
	BaseAmount     float64  `json:"base_amount"`
	OverrideAmount *float64 `json:"override_amount,omitempty"`
	TotalAmount    float64  `json:"total_amount"`
}

type AvailabilityDate struct {
	Date      string `json:"date"`
	Available int    `json:"available"`
}

type AvailabilityResponse struct {
	RoomTypeID     string             `json:"room_type_id"`
	Available      bool               `json:"available"`
	AvailableRooms int                `json:"available_rooms"`
	TotalPrice     float64            `json:"total_price"`
	Dates          []AvailabilityDate `json:"dates"`
}

type CreateReservationRequest struct {
	GuestName     string           `json:"guest_name"     validate:"required"`
	GuestEmail    string           `json:"guest_email"    validate:"required,email"`
	GuestPhone    string           `json:"guest_phone"`
	Bookings      []BookingRequest `json:"bookings"       validate:"required,min=1,max=4,dive"`
	PaymentMethod string           `json:"payment_method" validate:"required,oneof=card crypto"`
}

type BookingRequest struct {
	RoomID            string `json:"room_id"             validate:"required,uuid"`
	CheckIn           string `json:"check_in"            validate:"required"`
	CheckOut          string `json:"check_out"           validate:"required"`
	BookingType       string `json:"booking_type"        validate:"required,oneof=daily hourly"`
	ExpectedOccupants int    `json:"expected_occupants"  validate:"min=1"`
}

type BookingResponse struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	RoomTypeID  string    `json:"room_type_id"`
	BookingType string    `json:"booking_type"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Status      string    `json:"status"`
	Amount      float64   `json:"amount"`
}

type ReservationResponse struct {
	ID                 string            `json:"id"`
	ReferenceCode      string            `json:"reference_code"`
	GuestName          string            `json:"guest_name"`
	GuestEmail         string            `json:"guest_email"`
	GuestPhone         string            `json:"guest_phone"`
	TotalAmount        float64           `json:"total_amount"`
	Status             string            `json:"status"`
	CancellationReason string            `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	Bookings           []BookingResponse `json:"bookings,omitempty"`
}

type CancelRequest struct {
	Otp    string `json:"otp"    validate:"required"`
	Reason string `json:"reason"`
}

type RequestOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ProcessPaymentRequest struct {
	ReservationID string `json:"reservation_id" validate:"required,uuid"`
	Method        string `json:"method"         validate:"required,oneof=card crypto"`
}

type PaymentResponse struct {
	ID                string  `json:"id"`
	ReservationID     string  `json:"reservation_id"`
	Amount            float64 `json:"amount"`
	Provider          string  `json:"provider"`
	Status            string  `json:"status"`
	ProviderReference string  `json:"provider_reference,omitempty"`
}
