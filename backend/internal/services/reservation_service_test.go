package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

type mockOTPStore struct {
	data map[string]string
}

func (m *mockOTPStore) Set(ctx context.Context, key, otp string, ttl time.Duration) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = otp
	return nil
}

func (m *mockOTPStore) Get(ctx context.Context, key string) (string, error) {
	if m.data == nil {
		return "", nil
	}
	val, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockOTPStore) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

type mockReservationRepo struct {
	reservations map[uuid.UUID]*models.Reservation
	byReference  map[string]*models.Reservation
}

func (m *mockReservationRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, ErrReservationNotFound
	}
	return r, nil
}

func (m *mockReservationRepo) FindByReferenceCode(ctx context.Context, ref string) (*models.Reservation, error) {
	r, ok := m.byReference[ref]
	if !ok {
		return nil, ErrReservationNotFound
	}
	return r, nil
}

func (m *mockReservationRepo) FindAll(ctx context.Context, status string, from, to time.Time) ([]models.Reservation, error) {
	var out []models.Reservation
	for _, r := range m.reservations {
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockReservationRepo) Create(ctx context.Context, r *models.Reservation) error {
	m.reservations[r.ID] = r
	m.byReference[r.ReferenceCode] = r
	return nil
}

func (m *mockReservationRepo) Update(ctx context.Context, r *models.Reservation) error {
	m.reservations[r.ID] = r
	m.byReference[r.ReferenceCode] = r
	return nil
}

func (m *mockReservationRepo) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	var out []models.Reservation
	for _, r := range m.reservations {
		if r.CustomerID != nil && *r.CustomerID == customerID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *mockReservationRepo) FindByIdempotencyKey(ctx context.Context, key string) (*models.Reservation, error) {
	for _, r := range m.reservations {
		if r.IdempotencyKey != nil && *r.IdempotencyKey == key {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

type mockBookingRepo struct {
	bookings map[uuid.UUID]*models.Booking
}

func (m *mockBookingRepo) FindByReservationID(ctx context.Context, id uuid.UUID) ([]models.Booking, error) {
	var out []models.Booking
	for _, b := range m.bookings {
		if b.ReservationID == id {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (m *mockBookingRepo) Create(ctx context.Context, b *models.Booking) error {
	m.bookings[b.ID] = b
	return nil
}

func (m *mockBookingRepo) Update(ctx context.Context, b *models.Booking) error {
	m.bookings[b.ID] = b
	return nil
}

type mockPaymentRepo struct {
	payments map[uuid.UUID]*models.Payment
}

func (m *mockPaymentRepo) FindByReservationID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	return nil, nil
}

func (m *mockPaymentRepo) Create(ctx context.Context, p *models.Payment) error {
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepo) Update(ctx context.Context, p *models.Payment) error {
	m.payments[p.ID] = p
	return nil
}

func TestReservationService_Create(t *testing.T) {
	roomTypeID := uuid.New()
	roomID := uuid.New()

	roomRepo := &struct{ mockRoomRepo }{mockRoomRepo{rooms: map[uuid.UUID]*models.Room{
		roomID: {ID: roomID, RoomTypeID: roomTypeID, RoomNumber: "101", Status: "active"},
	}}}
	rtRepo := &struct{ mockRoomTypeRepo }{mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{
		roomTypeID: {ID: roomTypeID, Name: "Deluxe", BaseRateDaily: 200},
	}}}
	invRepo := &struct{ mockInventoryRepo }{mockInventoryRepo{}}
	pricingRepo := &struct{ mockPricingRepo }{mockPricingRepo{}}

	svc := NewReservationService(
		&mockReservationRepo{reservations: map[uuid.UUID]*models.Reservation{}, byReference: map[string]*models.Reservation{}},
		&mockBookingRepo{bookings: map[uuid.UUID]*models.Booking{}},
		&mockPaymentRepo{payments: map[uuid.UUID]*models.Payment{}},
		roomRepo,
		invRepo,
		pricingRepo,
		rtRepo,
		&mockOTPStore{},
		nil,
		nil,
		nil,
	)

	input := CreateReservationInput{
		GuestName:   "Alice",
		GuestEmail:  "alice@test.com",
		GuestPhone:  "+234800000000",
		Bookings: []BookingInput{
			{
				RoomID:            roomID,
				CheckIn:           time.Date(2026, 7, 15, 14, 0, 0, 0, time.UTC),
				CheckOut:          time.Date(2026, 7, 17, 11, 0, 0, 0, time.UTC),
				BookingType:       "daily",
				ExpectedOccupants: 2,
			},
		},
		PaymentMethod: "card",
	}

	res, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if res.ReferenceCode == "" {
		t.Fatal("expected reference")
	}
	if res.GuestName != "Alice" {
		t.Errorf("expected Alice, got %s", res.GuestName)
	}
	if res.Status != "pending" {
		t.Errorf("expected pending, got %s", res.Status)
	}
	if res.TotalAmount != 400 { // 2 nights * 200
		t.Errorf("expected 400, got %.2f", res.TotalAmount)
	}
}

func TestReservationService_Create_maxRoomsExceeded(t *testing.T) {
	svc := NewReservationService(
		&mockReservationRepo{},
		&mockBookingRepo{},
		&mockPaymentRepo{},
		nil, nil, nil, nil, nil, nil, nil, nil,
	)

	input := CreateReservationInput{
		GuestName:  "Alice",
		GuestEmail: "alice@test.com",
		Bookings:   make([]BookingInput, 5),
	}

	_, err := svc.Create(context.Background(), input)
	if err != ErrMaxRoomsExceeded {
		t.Fatalf("expected ErrMaxRoomsExceeded, got %v", err)
	}
}

func TestReservationService_Create_noBookings(t *testing.T) {
	svc := NewReservationService(
		&mockReservationRepo{},
		&mockBookingRepo{},
		&mockPaymentRepo{},
		nil, nil, nil, nil, &mockOTPStore{}, nil, nil, nil,
	)

	input := CreateReservationInput{
		GuestName:  "Alice",
		GuestEmail: "alice@test.com",
		Bookings:   []BookingInput{},
	}

	_, err := svc.Create(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for no bookings")
	}
}

func TestReservationService_Create_invalidRoom(t *testing.T) {
	roomRepo := &struct{ mockRoomRepo }{mockRoomRepo{rooms: map[uuid.UUID]*models.Room{}}}
	svc := NewReservationService(
		&mockReservationRepo{reservations: map[uuid.UUID]*models.Reservation{}, byReference: map[string]*models.Reservation{}},
		&mockBookingRepo{bookings: map[uuid.UUID]*models.Booking{}},
		&mockPaymentRepo{payments: map[uuid.UUID]*models.Payment{}},
		roomRepo,
		&struct{ mockInventoryRepo }{},
		&struct{ mockPricingRepo }{},
		&struct{ mockRoomTypeRepo }{},
		&mockOTPStore{}, nil, nil, nil,
	)

	input := CreateReservationInput{
		GuestName:  "Alice",
		GuestEmail: "alice@test.com",
		Bookings: []BookingInput{
			{
				RoomID:   uuid.New(),
				CheckIn:  time.Date(2026, 7, 15, 14, 0, 0, 0, time.UTC),
				CheckOut: time.Date(2026, 7, 17, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	_, err := svc.Create(context.Background(), input)
	if err != ErrRoomNotFound {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestReservationService_Create_invalidDates(t *testing.T) {
	roomID := uuid.New()
	roomRepo := &struct{ mockRoomRepo }{mockRoomRepo{rooms: map[uuid.UUID]*models.Room{
		roomID: {ID: roomID, Status: "active"},
	}}}
	svc := NewReservationService(
		&mockReservationRepo{},
		&mockBookingRepo{},
		&mockPaymentRepo{},
		roomRepo,
		&struct{ mockInventoryRepo }{},
		&struct{ mockPricingRepo }{},
		&struct{ mockRoomTypeRepo }{},
		&mockOTPStore{}, nil, nil, nil,
	)

	input := CreateReservationInput{
		GuestName:  "Alice",
		GuestEmail: "alice@test.com",
		Bookings: []BookingInput{
			{
				RoomID:   roomID,
				CheckIn:  time.Date(2026, 7, 17, 14, 0, 0, 0, time.UTC),
				CheckOut: time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	_, err := svc.Create(context.Background(), input)
	if err != ErrInvalidBookingDates {
		t.Fatalf("expected ErrInvalidBookingDates, got %v", err)
	}
}

func TestReservationService_Lookup(t *testing.T) {
	roomTypeID := uuid.New()
	roomID := uuid.New()

	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{},
	}
	svc := NewReservationService(
		reservationRepo,
		&mockBookingRepo{bookings: map[uuid.UUID]*models.Booking{}},
		&mockPaymentRepo{payments: map[uuid.UUID]*models.Payment{}},
		&mockRoomRepo{rooms: map[uuid.UUID]*models.Room{roomID: {ID: roomID, RoomTypeID: roomTypeID, Status: "active"}}},
		&mockInventoryRepo{},
		&mockPricingRepo{},
		&mockRoomTypeRepo{roomTypes: map[uuid.UUID]*models.RoomType{roomTypeID: {ID: roomTypeID, BaseRateDaily: 200}}},
		&mockOTPStore{}, nil, nil, nil,
	)

	created, _ := svc.Create(context.Background(), CreateReservationInput{
		GuestName:  "Alice",
		GuestEmail: "alice@test.com",
		GuestPhone: "+234800000000",
		Bookings: []BookingInput{
			{
				RoomID:            roomID,
				CheckIn:           time.Date(2026, 7, 15, 14, 0, 0, 0, time.UTC),
				CheckOut:          time.Date(2026, 7, 17, 11, 0, 0, 0, time.UTC),
				BookingType:       "daily",
				ExpectedOccupants: 2,
			},
		},
		PaymentMethod: "card",
	})

	found, err := svc.Lookup(context.Background(), created.ReferenceCode, "alice@test.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if found.ReferenceCode != created.ReferenceCode {
		t.Errorf("expected reference %s, got %s", created.ReferenceCode, found.ReferenceCode)
	}
}

func TestReservationService_Lookup_wrongEmail(t *testing.T) {
	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference: map[string]*models.Reservation{
			"HB-TEST": {ReferenceCode: "HB-TEST", GuestEmail: "alice@test.com"},
		},
	}
	svc := NewReservationService(reservationRepo, nil, nil, nil, nil, nil, nil, &mockOTPStore{}, nil, nil, nil)

	_, err := svc.Lookup(context.Background(), "HB-TEST", "bob@test.com")
	if err != ErrReservationNotFound {
		t.Fatalf("expected ErrReservationNotFound, got %v", err)
	}
}

func TestReservationService_Cancel_withValidOTP(t *testing.T) {
	otpStore := &mockOTPStore{}
	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{},
	}
	bookingRepo := &mockBookingRepo{
		bookings: map[uuid.UUID]*models.Booking{},
	}
	svc := NewReservationService(reservationRepo, bookingRepo, nil, nil, nil, nil, nil, otpStore, nil, nil, nil)

	// Seed a reservation
	reservationRepo.Create(context.Background(), &models.Reservation{
		ID:           uuid.New(),
		ReferenceCode: "HB-TEST",
		GuestEmail:    "alice@test.com",
		Status:        "confirmed",
	})

	err := svc.RequestCancelOTP(context.Background(), "HB-TEST", "alice@test.com")
	if err != nil {
		t.Fatalf("RequestCancelOTP: %v", err)
	}

	otp, _ := otpStore.Get(context.Background(), "cancel_otp:HB-TEST")
	if otp == "" {
		t.Fatal("expected OTP to be stored")
	}

	err = svc.Cancel(context.Background(), "HB-TEST", otp, "changed mind")
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	res, _ := reservationRepo.FindByReferenceCode(context.Background(), "HB-TEST")
	if res.Status != "cancelled" {
		t.Errorf("expected cancelled, got %s", res.Status)
	}
}

func TestReservationService_Cancel_wrongOTP(t *testing.T) {
	otpStore := &mockOTPStore{}
	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{
			"HB-TEST": {ReferenceCode: "HB-TEST", GuestEmail: "alice@test.com", Status: "confirmed"},
		},
	}
	svc := NewReservationService(reservationRepo, nil, nil, nil, nil, nil, nil, otpStore, nil, nil, nil)

	// Store a different OTP
	otpStore.Set(context.Background(), "cancel_otp:HB-TEST", "654321", 15*time.Minute)

	err := svc.Cancel(context.Background(), "HB-TEST", "000000", "nope")
	if err != ErrInvalidOTP {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
}

func TestReservationService_Cancel_noOTPRequested(t *testing.T) {
	otpStore := &mockOTPStore{}
	reservationRepo := &mockReservationRepo{
		byReference: map[string]*models.Reservation{
			"HB-TEST": {ReferenceCode: "HB-TEST", GuestEmail: "alice@test.com", Status: "confirmed"},
		},
	}
	svc := NewReservationService(reservationRepo, nil, nil, nil, nil, nil, nil, otpStore, nil, nil, nil)

	err := svc.Cancel(context.Background(), "HB-TEST", "000000", "nope")
	if err != ErrInvalidOTP {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
}

func TestReservationService_Cancel_alreadyCancelled(t *testing.T) {
	otpStore := &mockOTPStore{}
	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference: map[string]*models.Reservation{
			"HB-TEST": {ReferenceCode: "HB-TEST", GuestEmail: "a@b.com", Status: "cancelled"},
		},
	}
	svc := NewReservationService(reservationRepo, nil, nil, nil, nil, nil, nil, otpStore, nil, nil, nil)

	err := svc.RequestCancelOTP(context.Background(), "HB-TEST", "a@b.com")
	if err != ErrAlreadyCancelled {
		t.Fatalf("expected ErrAlreadyCancelled, got %v", err)
	}
}

func TestReservationService_RequestCancelOTP_wrongEmail(t *testing.T) {
	otpStore := &mockOTPStore{}
	reservationRepo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference: map[string]*models.Reservation{
			"HB-TEST": {ReferenceCode: "HB-TEST", GuestEmail: "alice@test.com"},
		},
	}
	svc := NewReservationService(reservationRepo, nil, nil, nil, nil, nil, nil, otpStore, nil, nil, nil)

	err := svc.RequestCancelOTP(context.Background(), "HB-TEST", "bob@test.com")
	if err != ErrReservationNotFound {
		t.Fatalf("expected ErrReservationNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// FindAll, FindByID, UpdateStatus tests
// ---------------------------------------------------------------------------

func TestReservationService_FindAll_empty(t *testing.T) {
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	results, err := svc.FindAll(context.Background(), "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestReservationService_FindAll_withData(t *testing.T) {
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{
			uuid.New(): {ID: uuid.New(), ReferenceCode: "HB-001", GuestName: "Alice", Status: "confirmed"},
			uuid.New(): {ID: uuid.New(), ReferenceCode: "HB-002", GuestName: "Bob", Status: "pending"},
		},
		byReference: map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	results, err := svc.FindAll(context.Background(), "", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestReservationService_FindByID_ok(t *testing.T) {
	resID := uuid.New()
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{
			resID: {ID: resID, ReferenceCode: "HB-001", GuestName: "Alice"},
		},
		byReference: map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	res, err := svc.FindByID(context.Background(), resID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if res.GuestName != "Alice" {
		t.Errorf("expected Alice, got %s", res.GuestName)
	}
}

func TestReservationService_FindByID_notFound(t *testing.T) {
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.FindByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent reservation")
	}
}

func TestReservationService_UpdateStatus_ok(t *testing.T) {
	resID := uuid.New()
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{
			resID: {ID: resID, ReferenceCode: "HB-001", Status: "pending"},
		},
		byReference: map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if err := svc.UpdateStatus(context.Background(), resID, "confirmed", ""); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if repo.reservations[resID].Status != "confirmed" {
		t.Errorf("expected confirmed, got %s", repo.reservations[resID].Status)
	}
}

func TestReservationService_UpdateStatus_notFound(t *testing.T) {
	repo := &mockReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{},
		byReference:  map[string]*models.Reservation{},
	}
	svc := NewReservationService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	err := svc.UpdateStatus(context.Background(), uuid.New(), "confirmed", "")
	if err == nil {
		t.Fatal("expected error for non-existent reservation")
	}
}

func TestReservationService_FindAllBookingsByReservation_ok(t *testing.T) {
	bookingID := uuid.New()
	resID := uuid.New()
	bookingRepo := &mockBookingRepo{
		bookings: map[uuid.UUID]*models.Booking{
			bookingID: {ID: bookingID, ReservationID: resID, RoomID: uuid.New()},
		},
	}
	svc := NewReservationService(nil, bookingRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	bookings, err := svc.FindAllBookingsByReservation(context.Background(), resID)
	if err != nil {
		t.Fatalf("FindAllBookingsByReservation: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}
}

func TestReservationService_FindAllBookingsByReservation_empty(t *testing.T) {
	bookingRepo := &mockBookingRepo{
		bookings: map[uuid.UUID]*models.Booking{},
	}
	svc := NewReservationService(nil, bookingRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	bookings, err := svc.FindAllBookingsByReservation(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("FindAllBookingsByReservation: %v", err)
	}
	if len(bookings) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(bookings))
	}
}
