package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

// mockBookingOps satisfies BookingOps.
// Named with the "Ops" suffix to avoid collision with mockBookingRepo in reservation_service_test.go,
// which has the same method set but is a separate package-level type.
// In Go, two types with identical method signatures are still distinct named types;
// however, to prevent duplicate-type compile errors when both files are in the same package,
// we use a different name here.
type mockBookingOps struct {
	bookings map[uuid.UUID]*models.Booking

	// capture the last Update call for CheckIn / CheckOut assertions
	lastUpdated *models.Booking
}

func (m *mockBookingOps) FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	var out []models.Booking
	for _, b := range m.bookings {
		if b.ReservationID == reservationID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (m *mockBookingOps) Create(ctx context.Context, b *models.Booking) error {
	if m.bookings == nil {
		m.bookings = make(map[uuid.UUID]*models.Booking)
	}
	m.bookings[b.ID] = b
	return nil
}

func (m *mockBookingOps) Update(ctx context.Context, b *models.Booking) error {
	m.lastUpdated = b
	if existing, ok := m.bookings[b.ID]; ok {
		existing.Status = b.Status
	} else {
		if m.bookings == nil {
			m.bookings = make(map[uuid.UUID]*models.Booking)
		}
		m.bookings[b.ID] = b
	}
	return nil
}

func (m *mockBookingOps) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.lastUpdated = &models.Booking{ID: id, Status: status}
	if existing, ok := m.bookings[id]; ok {
		existing.Status = status
	} else {
		if m.bookings == nil {
			m.bookings = make(map[uuid.UUID]*models.Booking)
		}
		m.bookings[id] = &models.Booking{ID: id, Status: status}
	}
	return nil
}

// mockRoomRepoForBooking satisfies RoomRepositoryService (FindByID only).
// Separate from mockRoomRepo in room_service_test.go which has additional methods
// (FindAll, Create, Update, Delete) satisfying the fuller RoomRepository interface.
type mockRoomRepoForBooking struct {
	rooms map[uuid.UUID]*models.Room
}

func (m *mockRoomRepoForBooking) FindByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	r, ok := m.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}
	return r, nil
}

func (m *mockRoomRepoForBooking) FindAll(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	out := make([]models.Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockRoomRepoForBooking) IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error) {
	return true, nil
}

// mockReservationCreator satisfies ReservationCreator.
type mockReservationCreator struct {
	reservations map[uuid.UUID]*models.Reservation
}

func (m *mockReservationCreator) FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (m *mockReservationCreator) Create(ctx context.Context, r *models.Reservation) error {
	if m.reservations == nil {
		m.reservations = make(map[uuid.UUID]*models.Reservation)
	}
	m.reservations[r.ID] = r
	return nil
}

// helpers

func newTestBookingService(
	bRepo *mockBookingOps,
	rRepo *mockRoomRepoForBooking,
	resRepo *mockReservationCreator,
) *BookingService {
	return NewBookingService(bRepo, rRepo, resRepo)
}

func baseWalkInInput(roomID uuid.UUID) WalkInBookingInput {
	return WalkInBookingInput{
		RoomID:            roomID,
		CheckIn:           time.Date(2026, 8, 1, 14, 0, 0, 0, time.UTC),
		CheckOut:          time.Date(2026, 8, 3, 11, 0, 0, 0, time.UTC),
		BookingType:       "daily",
		ExpectedOccupants: 2,
		Amount:            25000,
		GuestName:         "Bob",
		GuestEmail:        "bob@test.com",
		GuestPhone:        "+2348000001",
	}
}

func TestBookingService_CreateWalkIn_ok(t *testing.T) {
	roomID := uuid.New()
	rtID := uuid.New()
	roomRepo := &mockRoomRepoForBooking{
		rooms: map[uuid.UUID]*models.Room{
			roomID: {ID: roomID, RoomTypeID: rtID, RoomNumber: "205", Status: "active"},
		},
	}
	bookingRepo := &mockBookingOps{bookings: map[uuid.UUID]*models.Booking{}}
	resRepo := &mockReservationCreator{reservations: map[uuid.UUID]*models.Reservation{}}

	svc := newTestBookingService(bookingRepo, roomRepo, resRepo)

	res, err := svc.CreateWalkIn(context.Background(), baseWalkInInput(roomID))
	if err != nil {
		t.Fatalf("CreateWalkIn: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil reservation")
	}
	if res.Status != "confirmed" {
		t.Errorf("expected reservation status confirmed, got %s", res.Status)
	}
	if res.GuestName != "Bob" {
		t.Errorf("expected guest name Bob, got %s", res.GuestName)
	}
	if res.ReferenceCode == "" {
		t.Error("expected non-empty reference")
	}
	if res.TotalAmount != 25000 {
		t.Errorf("expected total amount 25000, got %.2f", res.TotalAmount)
	}

	// verify the booking was persisted
	if len(bookingRepo.bookings) != 1 {
		t.Errorf("expected 1 booking in repo, got %d", len(bookingRepo.bookings))
	}
	for _, b := range bookingRepo.bookings {
		if b.ReservationID != res.ID {
			t.Errorf("booking reservation_id mismatch: expected %s, got %s", res.ID, b.ReservationID)
		}
		if b.RoomID != roomID {
			t.Errorf("booking room_id mismatch: expected %s, got %s", roomID, b.RoomID)
		}
		if b.Status != "confirmed" {
			t.Errorf("expected booking status confirmed, got %s", b.Status)
		}
		if b.RoomTypeID != rtID {
			t.Errorf("expected booking room_type_id %s, got %s", rtID, b.RoomTypeID)
		}
		if b.Amount != 25000 {
			t.Errorf("expected booking amount 25000, got %.2f", b.Amount)
		}
	}
}

func TestBookingService_CreateWalkIn_roomNotFound(t *testing.T) {
	roomRepo := &mockRoomRepoForBooking{rooms: map[uuid.UUID]*models.Room{}} // empty — no rooms
	bookingRepo := &mockBookingOps{bookings: map[uuid.UUID]*models.Booking{}}
	resRepo := &mockReservationCreator{reservations: map[uuid.UUID]*models.Reservation{}}

	svc := newTestBookingService(bookingRepo, roomRepo, resRepo)

	_, err := svc.CreateWalkIn(context.Background(), baseWalkInInput(uuid.New()))
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestBookingService_CreateWalkIn_roomUnavailable(t *testing.T) {
	roomID := uuid.New()
	roomRepo := &mockRoomRepoForBooking{
		rooms: map[uuid.UUID]*models.Room{
			roomID: {ID: roomID, RoomNumber: "301", Status: "maintenance"},
		},
	}
	bookingRepo := &mockBookingOps{bookings: map[uuid.UUID]*models.Booking{}}
	resRepo := &mockReservationCreator{reservations: map[uuid.UUID]*models.Reservation{}}

	svc := newTestBookingService(bookingRepo, roomRepo, resRepo)

	_, err := svc.CreateWalkIn(context.Background(), baseWalkInInput(roomID))
	if err != ErrRoomNotAvailable {
		t.Fatalf("expected ErrRoomNotAvailable, got %v", err)
	}
}

func TestBookingService_CheckIn_ok(t *testing.T) {
	bookingID := uuid.New()
	bookingRepo := &mockBookingOps{
		bookings: map[uuid.UUID]*models.Booking{
			bookingID: {ID: bookingID, Status: "confirmed"},
		},
	}
	svc := newTestBookingService(bookingRepo, nil, nil)

	if err := svc.CheckIn(context.Background(), bookingID); err != nil {
		t.Fatalf("CheckIn: %v", err)
	}
	if bookingRepo.lastUpdated == nil {
		t.Fatal("expected Update to have been called")
	}
	if bookingRepo.lastUpdated.ID != bookingID {
		t.Errorf("expected booking ID %s, got %s", bookingID, bookingRepo.lastUpdated.ID)
	}
	if bookingRepo.lastUpdated.Status != "checked_in" {
		t.Errorf("expected status checked_in, got %s", bookingRepo.lastUpdated.Status)
	}
}

func TestBookingService_CheckOut_ok(t *testing.T) {
	bookingID := uuid.New()
	bookingRepo := &mockBookingOps{
		bookings: map[uuid.UUID]*models.Booking{
			bookingID: {ID: bookingID, Status: "checked_in"},
		},
	}
	svc := newTestBookingService(bookingRepo, nil, nil)

	if err := svc.CheckOut(context.Background(), bookingID); err != nil {
		t.Fatalf("CheckOut: %v", err)
	}
	if bookingRepo.lastUpdated == nil {
		t.Fatal("expected Update to have been called")
	}
	if bookingRepo.lastUpdated.ID != bookingID {
		t.Errorf("expected booking ID %s, got %s", bookingID, bookingRepo.lastUpdated.ID)
	}
	if bookingRepo.lastUpdated.Status != "checked_out" {
		t.Errorf("expected status checked_out, got %s", bookingRepo.lastUpdated.Status)
	}
}
