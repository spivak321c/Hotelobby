package services

import (
	"context"
	"errors"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

var (
	ErrBookingNotFound = errors.New("booking not found")
)

type BookingService struct {
	bookingRepo    BookingOps
	roomRepo       RoomRepositoryService
	reservationRepo ReservationCreator
}

type BookingOps interface {
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error)
	Create(ctx context.Context, b *models.Booking) error
	Update(ctx context.Context, b *models.Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type ReservationCreator interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error)
	Create(ctx context.Context, r *models.Reservation) error
}

type WalkInBookingInput struct {
	RoomID            uuid.UUID `json:"room_id"`
	CheckIn           time.Time `json:"check_in"`
	CheckOut          time.Time `json:"check_out"`
	BookingType       string    `json:"booking_type"`
	ExpectedOccupants int       `json:"expected_occupants"`
	Amount            float64   `json:"amount"`
	GuestName         string    `json:"guest_name"`
	GuestEmail        string    `json:"guest_email"`
	GuestPhone        string    `json:"guest_phone"`
}

func NewBookingService(bRepo BookingOps, rRepo RoomRepositoryService, resRepo ReservationCreator) *BookingService {
	return &BookingService{
		bookingRepo:      bRepo,
		roomRepo:         rRepo,
		reservationRepo:  resRepo,
	}
}

// Admin creates a walk-in booking — creates reservation + booking in confirmed status
func (s *BookingService) CreateWalkIn(ctx context.Context, input WalkInBookingInput) (*models.Reservation, error) {
	room, err := s.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	if room.Status != "active" {
		return nil, ErrRoomNotAvailable
	}

	reservation := &models.Reservation{
		ID:            uuid.New(),
		ReferenceCode: generateReference(),
		GuestName:     input.GuestName,
		GuestEmail:     input.GuestEmail,
		GuestPhone:     input.GuestPhone,
		TotalAmount:   input.Amount,
		Status:        "confirmed",
	}
	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, err
	}

	booking := &models.Booking{
		ID:          uuid.New(),
		ReservationID: reservation.ID,
		RoomID:      input.RoomID,
		RoomTypeID:  room.RoomTypeID,
		StartsAt:    input.CheckIn,
		EndsAt:      input.CheckOut,
		BookingType: input.BookingType,
		Amount:      input.Amount,
		Status:      "confirmed",
	}
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *BookingService) CheckIn(ctx context.Context, bookingID uuid.UUID) error {
	return s.bookingRepo.UpdateStatus(ctx, bookingID, "checked_in")
}

func (s *BookingService) CheckOut(ctx context.Context, bookingID uuid.UUID) error {
	return s.bookingRepo.UpdateStatus(ctx, bookingID, "checked_out")
}
