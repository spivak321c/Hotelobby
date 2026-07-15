package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/repositories"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Room repo mocks — satisfy repositories.* interfaces exactly
// ---------------------------------------------------------------------------

type roomTypeRepoMock struct{ types []models.RoomType }

func (r *roomTypeRepoMock) FindAll(_ context.Context) ([]models.RoomType, error) {
	return r.types, nil
}
func (r *roomTypeRepoMock) FindByID(_ context.Context, id uuid.UUID) (*models.RoomType, error) {
	for i := range r.types {
		if r.types[i].ID == id {
			return &r.types[i], nil
		}
	}
	return nil, errors.New("not found")
}
func (r *roomTypeRepoMock) Create(_ context.Context, rt *models.RoomType) error {
	r.types = append(r.types, *rt)
	return nil
}
func (r *roomTypeRepoMock) Update(_ context.Context, _ *models.RoomType) error { return nil }
func (r *roomTypeRepoMock) Delete(_ context.Context, _ uuid.UUID) error        { return nil }

var _ repositories.RoomTypeRepository = (*roomTypeRepoMock)(nil)

type roomRepoMock struct{ rooms []models.Room }

func (r *roomRepoMock) FindAll(_ context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	var out []models.Room
	for _, rm := range r.rooms {
		if roomTypeID != nil && rm.RoomTypeID != *roomTypeID {
			continue
		}
		if status != "" && rm.Status != status {
			continue
		}
		out = append(out, rm)
	}
	return out, nil
}
func (r *roomRepoMock) FindByID(_ context.Context, id uuid.UUID) (*models.Room, error) {
	for i := range r.rooms {
		if r.rooms[i].ID == id {
			return &r.rooms[i], nil
		}
	}
	return nil, errors.New("not found")
}
func (r *roomRepoMock) Create(_ context.Context, rm *models.Room) error {
	r.rooms = append(r.rooms, *rm)
	return nil
}
func (r *roomRepoMock) Update(_ context.Context, _ *models.Room) error { return nil }
func (r *roomRepoMock) Delete(_ context.Context, _ uuid.UUID) error    { return nil }
func (r *roomRepoMock) CountActiveBookings(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }
func (r *roomRepoMock) CountRooms(_ context.Context) (int, error) { return len(r.rooms), nil }
func (r *roomRepoMock) IsAvailable(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) { return true, nil }

var _ repositories.RoomRepository = (*roomRepoMock)(nil)

type roomPricingRepoMock struct{}

func (r *roomPricingRepoMock) FindAll(_ context.Context, _ *uuid.UUID) ([]models.RoomPricing, error) {
	return nil, nil
}
func (r *roomPricingRepoMock) FindByID(_ context.Context, _ uuid.UUID) (*models.RoomPricing, error) {
	return nil, errors.New("not found")
}
func (r *roomPricingRepoMock) FindByRoomTypeID(_ context.Context, _ uuid.UUID) ([]models.RoomPricing, error) {
	return nil, nil
}
func (r *roomPricingRepoMock) Create(_ context.Context, _ *models.RoomPricing) error { return nil }
func (r *roomPricingRepoMock) Update(_ context.Context, _ *models.RoomPricing) error { return nil }
func (r *roomPricingRepoMock) Delete(_ context.Context, _ uuid.UUID) error           { return nil }

var _ repositories.RoomPricingRepository = (*roomPricingRepoMock)(nil)

type inventoryRepoMock struct{}

func (r *inventoryRepoMock) FindByRoomTypeAndDate(_ context.Context, _ uuid.UUID, _ time.Time) (*models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *inventoryRepoMock) FindByRoomTypeAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *inventoryRepoMock) IncrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }
func (r *inventoryRepoMock) DecrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }
func (r *inventoryRepoMock) SetInventory(_ context.Context, _ uuid.UUID, _ time.Time, _, _ int) error {
	return nil
}

var _ repositories.RoomTypeInventoryRepository = (*inventoryRepoMock)(nil)

type roomImageRepoMock struct{}

func (r *roomImageRepoMock) FindByRoomID(_ context.Context, _ uuid.UUID) ([]models.RoomImage, error) {
	return nil, nil
}
func (r *roomImageRepoMock) Create(_ context.Context, _ *models.RoomImage) error         { return nil }
func (r *roomImageRepoMock) Delete(_ context.Context, _ uuid.UUID) error                 { return nil }
func (r *roomImageRepoMock) SetPrimary(_ context.Context, _ uuid.UUID) error             { return nil }
func (r *roomImageRepoMock) Reorder(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error { return nil }

var _ repositories.RoomImageRepository = (*roomImageRepoMock)(nil)

// decodeSlice reads a bare JSON array from a response body.
func decodeSlice(t *testing.T, resp *http.Response, dest interface{}) error {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

// ---------------------------------------------------------------------------
// Room app factory
// ---------------------------------------------------------------------------

func newRoomApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()

	rtID := uuid.New()
	roomID := uuid.New()

	rtRepo := &roomTypeRepoMock{types: []models.RoomType{
		{ID: rtID, Name: "Deluxe", BaseRateHourly: 50, BaseRateDaily: 300},
	}}
	rRepo := &roomRepoMock{rooms: []models.Room{
		{ID: roomID, RoomTypeID: rtID, RoomNumber: "101", Status: "available"},
	}}

	roomSvc := services.NewRoomService(rtRepo, rRepo, &roomPricingRepoMock{}, &inventoryRepoMock{}, &roomImageRepoMock{})
	h := handlers.NewRoomHandler(roomSvc)

	app := fiber.New()
	app.Get("/api/room-types", h.ListRoomTypes)
	app.Get("/api/room-types/:id", h.GetRoomType)
	app.Get("/api/rooms", h.ListRooms)
	app.Get("/api/rooms/:id", h.GetRoom)
	return app, roomID
}

// ---------------------------------------------------------------------------
// Room handler tests
// ---------------------------------------------------------------------------

func TestRoomHandler_ListRoomTypes(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/room-types", nil), 5000)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var arr []map[string]interface{}
	if err := decodeSlice(t, resp, &arr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 room type, got %d", len(arr))
	}
	if arr[0]["name"] != "Deluxe" {
		t.Fatalf("expected Deluxe, got %v", arr[0]["name"])
	}
}

func TestRoomHandler_GetRoomType_notFound(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/room-types/"+uuid.New().String(), nil), 5000)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRoomHandler_GetRoomType_invalidID(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/room-types/not-a-uuid", nil), 5000)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRoomHandler_ListRooms_all(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/rooms", nil), 5000)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var arr []map[string]interface{}
	decodeSlice(t, resp, &arr)
	if len(arr) != 1 {
		t.Fatalf("expected 1 room, got %d", len(arr))
	}
}

func TestRoomHandler_ListRooms_filterByStatus(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/rooms?status=maintenance", nil), 5000)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var arr []map[string]interface{}
	decodeSlice(t, resp, &arr)
	if len(arr) != 0 {
		t.Fatalf("expected 0 rooms after filter, got %d", len(arr))
	}
}

func TestRoomHandler_GetRoom_ok(t *testing.T) {
	app, roomID := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/rooms/"+roomID.String(), nil), 5000)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRoomHandler_GetRoom_notFound(t *testing.T) {
	app, _ := newRoomApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/rooms/"+uuid.New().String(), nil), 5000)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
