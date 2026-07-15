package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Reservation-service repo mocks
// All satisfy the narrow interfaces declared in reservation_interfaces.go.
// ---------------------------------------------------------------------------

// resRepoMock satisfies services.ReservationRepository.
type resRepoMock struct {
	byRef map[string]*models.Reservation
	byID  map[uuid.UUID]*models.Reservation
}

func newResRepoMock() *resRepoMock {
	return &resRepoMock{
		byRef: make(map[string]*models.Reservation),
		byID:  make(map[uuid.UUID]*models.Reservation),
	}
}

func (r *resRepoMock) FindByID(_ context.Context, id uuid.UUID) (*models.Reservation, error) {
	if res, ok := r.byID[id]; ok {
		return res, nil
	}
	return nil, errors.New("not found")
}
func (r *resRepoMock) FindByReferenceCode(_ context.Context, ref string) (*models.Reservation, error) {
	if res, ok := r.byRef[ref]; ok {
		return res, nil
	}
	return nil, errors.New("not found")
}
func (r *resRepoMock) FindAll(_ context.Context, _ string, _, _ time.Time) ([]models.Reservation, error) {
	return nil, nil
}
func (r *resRepoMock) Create(_ context.Context, res *models.Reservation) error {
	r.byRef[res.ReferenceCode] = res
	r.byID[res.ID] = res
	return nil
}
func (r *resRepoMock) Update(_ context.Context, res *models.Reservation) error {
	r.byRef[res.ReferenceCode] = res
	r.byID[res.ID] = res
	return nil
}
func (r *resRepoMock) FindByCustomerID(_ context.Context, _ uuid.UUID) ([]models.Reservation, error) {
	return nil, nil
}
func (r *resRepoMock) FindByIdempotencyKey(_ context.Context, _ string) (*models.Reservation, error) {
	return nil, errors.New("not found")
}

// bookingRepoMock satisfies services.BookingRepository (reservation service).
type bookingRepoMock struct{}

func (r *bookingRepoMock) FindByReservationID(_ context.Context, _ uuid.UUID) ([]models.Booking, error) {
	return nil, nil
}
func (r *bookingRepoMock) Create(_ context.Context, _ *models.Booking) error { return nil }
func (r *bookingRepoMock) Update(_ context.Context, _ *models.Booking) error { return nil }

// paymentRepoForResMock satisfies services.PaymentRepository (reservation service).
type paymentRepoForResMock struct{}

func (r *paymentRepoForResMock) FindByReservationID(_ context.Context, _ uuid.UUID) (*models.Payment, error) {
	return nil, nil
}
func (r *paymentRepoForResMock) Create(_ context.Context, _ *models.Payment) error { return nil }
func (r *paymentRepoForResMock) Update(_ context.Context, _ *models.Payment) error { return nil }

// roomRepoForResMock satisfies services.RoomRepositoryService.
type roomRepoForResMock struct {
	rooms map[uuid.UUID]*models.Room
}

func (r *roomRepoForResMock) FindByID(_ context.Context, id uuid.UUID) (*models.Room, error) {
	if rm, ok := r.rooms[id]; ok {
		return rm, nil
	}
	return nil, errors.New("not found")
}
func (r *roomRepoForResMock) FindAll(_ context.Context, _ *uuid.UUID, _ string) ([]models.Room, error) {
	out := make([]models.Room, 0, len(r.rooms))
	for _, rm := range r.rooms {
		out = append(out, *rm)
	}
	return out, nil
}
func (r *roomRepoForResMock) IsAvailable(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) { return true, nil }

// invRepoForResMock satisfies services.InventoryRepositoryService.
type invRepoForResMock struct{}

func (r *invRepoForResMock) FindByRoomTypeAndDate(_ context.Context, _ uuid.UUID, _ time.Time) (*models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *invRepoForResMock) FindByRoomTypeAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *invRepoForResMock) IncrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }
func (r *invRepoForResMock) DecrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }

// pricingRepoForResMock satisfies services.PricingRepositoryService.
type pricingRepoForResMock struct{}

func (r *pricingRepoForResMock) FindByRoomTypeID(_ context.Context, _ uuid.UUID) ([]models.RoomPricing, error) {
	return nil, nil
}

// rtRepoForResMock satisfies services.RoomTypeRepositoryService.
type rtRepoForResMock struct {
	types map[uuid.UUID]*models.RoomType
}

func (r *rtRepoForResMock) FindByID(_ context.Context, id uuid.UUID) (*models.RoomType, error) {
	if rt, ok := r.types[id]; ok {
		return rt, nil
	}
	return nil, errors.New("not found")
}

// ---------------------------------------------------------------------------
// Reservation app factory
// ---------------------------------------------------------------------------

func newReservationApp(t *testing.T) *fiber.App {
	t.Helper()

	resSvc := services.NewReservationService(
		newResRepoMock(),
		&bookingRepoMock{},
		&paymentRepoForResMock{},
		&roomRepoForResMock{rooms: map[uuid.UUID]*models.Room{}},
		&invRepoForResMock{},
		&pricingRepoForResMock{},
		&rtRepoForResMock{types: map[uuid.UUID]*models.RoomType{}},
		services.NewMemoryOTPStore(),
		nil, // email service
		nil, // sse hub
		nil, // inventory svc
	)
	h := handlers.NewReservationHandler(resSvc)

	app := fiber.New()
	app.Post("/api/reservations", h.Create)
	app.Get("/api/reservations/:reference", h.Lookup)
	app.Post("/api/reservations/:reference/cancel/otp", h.RequestCancelOTP)
	app.Post("/api/reservations/:reference/cancel", h.Cancel)
	return app
}

// ---------------------------------------------------------------------------
// Reservation handler tests — input validation and routing
// Business logic is covered fully in services/reservation_service_test.go.
// ---------------------------------------------------------------------------

func TestReservationHandler_Create_missingGuestName(t *testing.T) {
	app := newReservationApp(t)
	resp := doJSON(t, app, "POST", "/api/reservations", map[string]interface{}{
		"guest_email":    "alice@example.com",
		"payment_method": "card",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReservationHandler_Create_invalidRoomID(t *testing.T) {
	app := newReservationApp(t)
	resp := doJSON(t, app, "POST", "/api/reservations", map[string]interface{}{
		"guest_name":     "Alice",
		"guest_email":    "alice@example.com",
		"payment_method": "card",
		"bookings": []map[string]interface{}{
			{
				"room_id":      "not-a-uuid",
				"check_in":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"check_out":    time.Now().Add(48 * time.Hour).Format(time.RFC3339),
				"booking_type": "daily",
			},
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReservationHandler_Lookup_missingEmail(t *testing.T) {
	app := newReservationApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/reservations/REF001", nil), 5000)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReservationHandler_Lookup_notFound(t *testing.T) {
	app := newReservationApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/reservations/UNKNOWN?email=x@x.com", nil), 5000)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestReservationHandler_Cancel_missingOTP(t *testing.T) {
	app := newReservationApp(t)
	resp := doJSON(t, app, "POST", "/api/reservations/REF001/cancel", map[string]string{
		"reason": "changed my mind",
		// otp intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReservationHandler_RequestCancelOTP_missingEmail(t *testing.T) {
	app := newReservationApp(t)
	resp := doJSON(t, app, "POST", "/api/reservations/REF001/cancel/otp", map[string]string{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
