package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"
	"hotel_lobby/internal/sse"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Full repository mocks for AdminHandler (satisfy repositories.* interfaces)
// ---------------------------------------------------------------------------

type admRoomTypeRepo struct {
	types map[uuid.UUID]*models.RoomType
}

func newAdmRoomTypeRepo() *admRoomTypeRepo {
	return &admRoomTypeRepo{types: make(map[uuid.UUID]*models.RoomType)}
}

func (r *admRoomTypeRepo) FindAll(_ context.Context) ([]models.RoomType, error) {
	out := make([]models.RoomType, 0, len(r.types))
	for _, rt := range r.types {
		out = append(out, *rt)
	}
	return out, nil
}
func (r *admRoomTypeRepo) FindByID(_ context.Context, id uuid.UUID) (*models.RoomType, error) {
	if rt, ok := r.types[id]; ok {
		return rt, nil
	}
	return nil, errors.New("not found")
}
func (r *admRoomTypeRepo) Create(_ context.Context, rt *models.RoomType) error {
	r.types[rt.ID] = rt
	return nil
}
func (r *admRoomTypeRepo) Update(_ context.Context, rt *models.RoomType) error {
	r.types[rt.ID] = rt
	return nil
}
func (r *admRoomTypeRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.types, id)
	return nil
}

type admRoomRepo struct {
	rooms map[uuid.UUID]*models.Room
}

func newAdmRoomRepo() *admRoomRepo {
	return &admRoomRepo{rooms: make(map[uuid.UUID]*models.Room)}
}

func (r *admRoomRepo) FindAll(_ context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	var out []models.Room
	for _, rm := range r.rooms {
		if roomTypeID != nil && rm.RoomTypeID != *roomTypeID {
			continue
		}
		if status != "" && rm.Status != status {
			continue
		}
		out = append(out, *rm)
	}
	return out, nil
}
func (r *admRoomRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Room, error) {
	if rm, ok := r.rooms[id]; ok {
		return rm, nil
	}
	return nil, errors.New("not found")
}
func (r *admRoomRepo) Create(_ context.Context, rm *models.Room) error {
	r.rooms[rm.ID] = rm
	return nil
}
func (r *admRoomRepo) Update(_ context.Context, rm *models.Room) error {
	r.rooms[rm.ID] = rm
	return nil
}
func (r *admRoomRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.rooms, id)
	return nil
}
func (r *admRoomRepo) CountActiveBookings(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }
func (r *admRoomRepo) CountRooms(_ context.Context) (int, error) { return len(r.rooms), nil }
func (r *admRoomRepo) IsAvailable(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) { return true, nil }

type admPricingRepo struct {
	prices map[uuid.UUID]*models.RoomPricing
}

func newAdmPricingRepo() *admPricingRepo {
	return &admPricingRepo{prices: make(map[uuid.UUID]*models.RoomPricing)}
}

func (r *admPricingRepo) FindAll(_ context.Context, roomTypeID *uuid.UUID) ([]models.RoomPricing, error) {
	var out []models.RoomPricing
	for _, p := range r.prices {
		if roomTypeID != nil && p.RoomTypeID != *roomTypeID {
			continue
		}
		out = append(out, *p)
	}
	return out, nil
}
func (r *admPricingRepo) FindByID(_ context.Context, id uuid.UUID) (*models.RoomPricing, error) {
	if p, ok := r.prices[id]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}
func (r *admPricingRepo) FindByRoomTypeID(_ context.Context, rtID uuid.UUID) ([]models.RoomPricing, error) {
	var out []models.RoomPricing
	for _, p := range r.prices {
		if p.RoomTypeID == rtID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (r *admPricingRepo) Create(_ context.Context, p *models.RoomPricing) error {
	r.prices[p.ID] = p
	return nil
}
func (r *admPricingRepo) Update(_ context.Context, p *models.RoomPricing) error {
	r.prices[p.ID] = p
	return nil
}
func (r *admPricingRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.prices, id)
	return nil
}

type admInventoryRepo struct {
	records map[string]*models.RoomTypeInventory
}

func newAdmInventoryRepo() *admInventoryRepo {
	return &admInventoryRepo{records: make(map[string]*models.RoomTypeInventory)}
}

func invKey(rtID uuid.UUID, date time.Time) string {
	return fmt.Sprintf("%s_%s", rtID.String(), date.Format("2006-01-02"))
}

func (r *admInventoryRepo) FindByRoomTypeAndDate(_ context.Context, rtID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error) {
	if inv, ok := r.records[invKey(rtID, date)]; ok {
		return inv, nil
	}
	return nil, errors.New("not found")
}
func (r *admInventoryRepo) FindByRoomTypeAndDateRange(_ context.Context, rtID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error) {
	var out []models.RoomTypeInventory
	for _, inv := range r.records {
		if inv.RoomTypeID == rtID && !inv.Date.Before(from) && !inv.Date.After(to) {
			out = append(out, *inv)
		}
	}
	return out, nil
}
func (r *admInventoryRepo) IncrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}
func (r *admInventoryRepo) DecrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}
func (r *admInventoryRepo) SetInventory(_ context.Context, rtID uuid.UUID, date time.Time, totalRooms, bookedRooms int) error {
	r.records[invKey(rtID, date)] = &models.RoomTypeInventory{
		RoomTypeID:  rtID,
		Date:        date,
		TotalRooms:  totalRooms,
		BookedRooms: bookedRooms,
	}
	return nil
}

type admImageRepo struct {
	images map[uuid.UUID][]models.RoomImage
}

func newAdmImageRepo() *admImageRepo {
	return &admImageRepo{images: make(map[uuid.UUID][]models.RoomImage)}
}

func (r *admImageRepo) FindByRoomID(_ context.Context, roomID uuid.UUID) ([]models.RoomImage, error) {
	return r.images[roomID], nil
}
func (r *admImageRepo) Create(_ context.Context, img *models.RoomImage) error    { return nil }
func (r *admImageRepo) Delete(_ context.Context, _ uuid.UUID) error             { return nil }
func (r *admImageRepo) SetPrimary(_ context.Context, _ uuid.UUID) error         { return nil }
func (r *admImageRepo) Reorder(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error { return nil }

type admCustomerRepo struct {
	customers map[uuid.UUID]*models.Customer
}

func newAdmCustomerRepo() *admCustomerRepo {
	return &admCustomerRepo{customers: make(map[uuid.UUID]*models.Customer)}
}

func (r *admCustomerRepo) FindAll(_ context.Context) ([]models.Customer, error) {
	out := make([]models.Customer, 0, len(r.customers))
	for _, c := range r.customers {
		out = append(out, *c)
	}
	return out, nil
}
func (r *admCustomerRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Customer, error) {
	if c, ok := r.customers[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}
func (r *admCustomerRepo) FindByEmail(_ context.Context, email string) (*models.Customer, error) {
	for _, c := range r.customers {
		if c.Email == email {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *admCustomerRepo) Create(_ context.Context, c *models.Customer) error {
	r.customers[c.ID] = c
	return nil
}
func (r *admCustomerRepo) Update(_ context.Context, c *models.Customer) error {
	r.customers[c.ID] = c
	return nil
}
func (r *admCustomerRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.customers, id)
	return nil
}

type admAdminRepo struct {
	admins map[uuid.UUID]*models.Admin
}

func newAdmAdminRepo() *admAdminRepo {
	return &admAdminRepo{admins: make(map[uuid.UUID]*models.Admin)}
}

func (r *admAdminRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Admin, error) {
	if a, ok := r.admins[id]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}
func (r *admAdminRepo) FindByEmail(_ context.Context, email string) (*models.Admin, error) {
	for _, a := range r.admins {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *admAdminRepo) FindAll(_ context.Context) ([]models.Admin, error) {
	out := make([]models.Admin, 0, len(r.admins))
	for _, a := range r.admins {
		out = append(out, *a)
	}
	return out, nil
}
func (r *admAdminRepo) Create(_ context.Context, a *models.Admin) error {
	r.admins[a.ID] = a
	return nil
}
func (r *admAdminRepo) Update(_ context.Context, a *models.Admin) error {
	r.admins[a.ID] = a
	return nil
}
func (r *admAdminRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.admins, id)
	return nil
}

// ---------------------------------------------------------------------------
// Service-level mocks for reservation service
// ---------------------------------------------------------------------------

type admResRepo struct {
	reservations map[uuid.UUID]*models.Reservation
	byRef        map[string]*models.Reservation
}

func newAdmResRepo() *admResRepo {
	return &admResRepo{
		reservations: make(map[uuid.UUID]*models.Reservation),
		byRef:        make(map[string]*models.Reservation),
	}
}

func (r *admResRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Reservation, error) {
	if res, ok := r.reservations[id]; ok {
		return res, nil
	}
	return nil, errors.New("not found")
}
func (r *admResRepo) FindByReferenceCode(_ context.Context, ref string) (*models.Reservation, error) {
	if res, ok := r.byRef[ref]; ok {
		return res, nil
	}
	return nil, errors.New("not found")
}
func (r *admResRepo) FindAll(_ context.Context, _ string, _, _ time.Time) ([]models.Reservation, error) {
	out := make([]models.Reservation, 0, len(r.reservations))
	for _, res := range r.reservations {
		out = append(out, *res)
	}
	return out, nil
}
func (r *admResRepo) Create(_ context.Context, res *models.Reservation) error {
	r.reservations[res.ID] = res
	r.byRef[res.ReferenceCode] = res
	return nil
}
func (r *admResRepo) Update(_ context.Context, res *models.Reservation) error {
	r.reservations[res.ID] = res
	r.byRef[res.ReferenceCode] = res
	return nil
}
func (r *admResRepo) FindByCustomerID(_ context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	var out []models.Reservation
	for _, res := range r.reservations {
		if res.CustomerID != nil && *res.CustomerID == customerID {
			out = append(out, *res)
		}
	}
	return out, nil
}
func (r *admResRepo) FindByIdempotencyKey(_ context.Context, _ string) (*models.Reservation, error) {
	return nil, errors.New("not found")
}

type admBookingRepo struct{}

func (r *admBookingRepo) FindByReservationID(_ context.Context, _ uuid.UUID) ([]models.Booking, error) {
	return nil, nil
}
func (r *admBookingRepo) Create(_ context.Context, _ *models.Booking) error { return nil }
func (r *admBookingRepo) Update(_ context.Context, _ *models.Booking) error { return nil }
func (r *admBookingRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ string) error { return nil }

type admPaymentRepo struct{}

func (r *admPaymentRepo) FindByReservationID(_ context.Context, _ uuid.UUID) (*models.Payment, error) {
	return nil, nil
}
func (r *admPaymentRepo) Create(_ context.Context, _ *models.Payment) error { return nil }
func (r *admPaymentRepo) Update(_ context.Context, _ *models.Payment) error { return nil }

type admRoomSvcRepo struct{ rooms map[uuid.UUID]*models.Room }

func (r *admRoomSvcRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Room, error) {
	if rm, ok := r.rooms[id]; ok {
		return rm, nil
	}
	return nil, errors.New("not found")
}
func (r *admRoomSvcRepo) IsAvailable(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) { return true, nil }

type admInvSvcRepo struct{}

func (r *admInvSvcRepo) FindByRoomTypeAndDate(_ context.Context, _ uuid.UUID, _ time.Time) (*models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *admInvSvcRepo) FindByRoomTypeAndDateRange(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]models.RoomTypeInventory, error) {
	return nil, nil
}
func (r *admInvSvcRepo) IncrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }
func (r *admInvSvcRepo) DecrementBooked(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }

type admPricingSvcRepo struct{}

func (r *admPricingSvcRepo) FindByRoomTypeID(_ context.Context, _ uuid.UUID) ([]models.RoomPricing, error) {
	return nil, nil
}

type admRTSvcRepo struct {
	types map[uuid.UUID]*models.RoomType
}

func (r *admRTSvcRepo) FindByID(_ context.Context, id uuid.UUID) (*models.RoomType, error) {
	if rt, ok := r.types[id]; ok {
		return rt, nil
	}
	return nil, errors.New("not found")
}

// ---------------------------------------------------------------------------
// AdminHandler wiring helper
// ---------------------------------------------------------------------------

type adminRepos struct {
	rtRepo   *admRoomTypeRepo
	roomRepo *admRoomRepo
	priceRepo *admPricingRepo
	invRepo  *admInventoryRepo
	imgRepo  *admImageRepo
	custRepo *admCustomerRepo
	adminRepo *admAdminRepo
	resRepo  *admResRepo
}

func setupAdminHandler(t *testing.T) (*fiber.App, *adminRepos) {
	t.Helper()

	rtRepo := newAdmRoomTypeRepo()
	roomRepo := newAdmRoomRepo()
	priceRepo := newAdmPricingRepo()
	invRepo := newAdmInventoryRepo()
	imgRepo := newAdmImageRepo()
	custRepo := newAdmCustomerRepo()
	adminRepo := newAdmAdminRepo()
	resRepo := newAdmResRepo()

	roomSvc := services.NewRoomService(rtRepo, roomRepo, priceRepo, invRepo, imgRepo)
	resSvc := services.NewReservationService(
		resRepo,
		&admBookingRepo{},
		&admPaymentRepo{},
		&admRoomSvcRepo{rooms: roomRepo.rooms},
		&admInvSvcRepo{},
		&admPricingSvcRepo{},
		&admRTSvcRepo{types: rtRepo.types},
		services.NewMemoryOTPStore(),
		nil,
		nil,
		nil,
	)
	bookingSvc := services.NewBookingService(&admBookingRepo{}, &admRoomSvcRepo{rooms: roomRepo.rooms}, resRepo)
	authSvc := services.NewAuthService(custRepo, adminRepo, "test-secret")
	imageSvc := services.NewImageService(nil, imgRepo)
	hub := sse.NewHub()

	h := handlers.NewAdminHandler(
		roomSvc, resSvc, bookingSvc, nil,
		authSvc, imageSvc, custRepo, adminRepo, hub,
	)

	app := fiber.New()
	// Room Types
	app.Get("/api/admin/room-types", h.ListRoomTypes)
	app.Post("/api/admin/room-types", h.CreateRoomType)
	app.Put("/api/admin/room-types/:id", h.UpdateRoomType)
	app.Delete("/api/admin/room-types/:id", h.DeleteRoomType)
	// Rooms
	app.Get("/api/admin/rooms", h.ListRooms)
	app.Post("/api/admin/rooms", h.CreateRoom)
	app.Put("/api/admin/rooms/:id", h.UpdateRoom)
	app.Delete("/api/admin/rooms/:id", h.DeleteRoom)
	// Pricing
	app.Get("/api/admin/room-pricing", h.ListRoomPricing)
	app.Post("/api/admin/room-pricing", h.CreateRoomPricing)
	app.Put("/api/admin/room-pricing/:id", h.UpdateRoomPricing)
	app.Delete("/api/admin/room-pricing/:id", h.DeleteRoomPricing)
	// Inventory
	app.Get("/api/admin/inventory", h.GetInventory)
	app.Put("/api/admin/inventory", h.UpdateInventory)
	// Reservations
	app.Get("/api/admin/reservations", h.ListReservations)
	app.Get("/api/admin/reservations/:id", h.GetReservation)
	app.Put("/api/admin/reservations/:id/status", h.UpdateReservationStatus)
	// Customers
	app.Get("/api/admin/customers", h.ListCustomers)
	app.Get("/api/admin/customers/:id", h.GetCustomer)
	// Admins
	app.Get("/api/admin/admins", h.ListAdmins)
	app.Post("/api/admin/admins", h.CreateAdmin)
	app.Put("/api/admin/admins/:id", h.UpdateAdmin)
	app.Delete("/api/admin/admins/:id", h.DeleteAdmin)
	// Reports
	app.Get("/api/admin/reports/bookings", h.BookingReport)
	app.Get("/api/admin/reports/occupancy", h.OccupancyReport)
	app.Get("/api/admin/reports/revenue", h.RevenueReport)

	repos := &adminRepos{
		rtRepo: rtRepo, roomRepo: roomRepo, priceRepo: priceRepo,
		invRepo: invRepo, imgRepo: imgRepo, custRepo: custRepo,
		adminRepo: adminRepo, resRepo: resRepo,
	}
	return app, repos
}

// ---------------------------------------------------------------------------
// Room Type tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListRoomTypes_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/room-types", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	if body["success"] != true {
		t.Fatal("expected success=true")
	}
}

func TestAdminHandler_ListRoomTypes_withData(t *testing.T) {
	app, repos := setupAdminHandler(t)
	id := uuid.New()
	repos.rtRepo.types[id] = &models.RoomType{ID: id, Name: "Deluxe", BaseRateHourly: 50, BaseRateDaily: 300}

	resp := doJSON(t, app, "GET", "/api/admin/room-types", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoomType_ok(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/room-types", map[string]interface{}{
		"name":              "Presidential",
		"description":       "Luxury suite",
		"base_rate_hourly":  100,
		"base_rate_daily":   800,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["name"] != "Presidential" {
		t.Fatalf("expected Presidential, got %v", data["name"])
	}
}

func TestAdminHandler_CreateRoomType_missingName(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/room-types", map[string]interface{}{
		"description": "no name",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateRoomType_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	id := uuid.New()
	repos.rtRepo.types[id] = &models.RoomType{ID: id, Name: "Old"}

	newName := "Updated"
	resp := doJSON(t, app, "PUT", "/api/admin/room-types/"+id.String(), map[string]interface{}{
		"name": newName,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["name"] != "Updated" {
		t.Fatalf("expected Updated, got %v", data["name"])
	}
}

func TestAdminHandler_UpdateRoomType_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/room-types/not-a-uuid", map[string]interface{}{
		"name": "x",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteRoomType_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	id := uuid.New()
	repos.rtRepo.types[id] = &models.RoomType{ID: id, Name: "ToDelete"}

	resp := doJSON(t, app, "DELETE", "/api/admin/room-types/"+id.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if _, ok := repos.rtRepo.types[id]; ok {
		t.Fatal("expected room type to be deleted")
	}
}

func TestAdminHandler_DeleteRoomType_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "DELETE", "/api/admin/room-types/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Room tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListRooms_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/rooms", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoom_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()
	repos.rtRepo.types[rtID] = &models.RoomType{ID: rtID, Name: "Standard"}

	resp := doJSON(t, app, "POST", "/api/admin/rooms", map[string]interface{}{
		"room_type_id": rtID.String(),
		"room_number":  "101",
		"status":       "available",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["room_number"] != "101" {
		t.Fatalf("expected 101, got %v", data["room_number"])
	}
}

func TestAdminHandler_CreateRoom_missingTypeID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/rooms", map[string]interface{}{
		"room_number": "101",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoom_missingRoomNumber(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()
	repos.rtRepo.types[rtID] = &models.RoomType{ID: rtID, Name: "Standard"}

	resp := doJSON(t, app, "POST", "/api/admin/rooms", map[string]interface{}{
		"room_type_id": rtID.String(),
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoom_invalidTypeID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/rooms", map[string]interface{}{
		"room_type_id": "bad",
		"room_number":  "101",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateRoom_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	roomID := uuid.New()
	repos.roomRepo.rooms[roomID] = &models.Room{ID: roomID, RoomNumber: "101", Status: "available"}

	newStatus := "maintenance"
	resp := doJSON(t, app, "PUT", "/api/admin/rooms/"+roomID.String(), map[string]interface{}{
		"status": newStatus,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateRoom_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/rooms/not-a-uuid", map[string]interface{}{
		"status": "x",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteRoom_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	roomID := uuid.New()
	repos.roomRepo.rooms[roomID] = &models.Room{ID: roomID, RoomNumber: "101"}

	resp := doJSON(t, app, "DELETE", "/api/admin/rooms/"+roomID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteRoom_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "DELETE", "/api/admin/rooms/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Pricing tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListRoomPricing_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/room-pricing", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_ListRoomPricing_invalidFilter(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/room-pricing?room_type_id=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoomPricing_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()
	repos.rtRepo.types[rtID] = &models.RoomType{ID: rtID, Name: "Standard"}

	resp := doJSON(t, app, "POST", "/api/admin/room-pricing", map[string]interface{}{
		"room_type_id":  rtID.String(),
		"rate_type":     "daily",
		"rate":          350.0,
		"effective_from": "2026-08-01",
		"effective_to":   "2026-08-31",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoomPricing_missingFields(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/room-pricing", map[string]interface{}{
		"rate": 100,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateRoomPricing_invalidDate(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()
	repos.rtRepo.types[rtID] = &models.RoomType{ID: rtID, Name: "Standard"}

	resp := doJSON(t, app, "POST", "/api/admin/room-pricing", map[string]interface{}{
		"room_type_id":  rtID.String(),
		"rate_type":     "daily",
		"effective_from": "not-a-date",
		"effective_to":   "2026-08-31",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateRoomPricing_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rpID := uuid.New()
	rtID := uuid.New()
	repos.priceRepo.prices[rpID] = &models.RoomPricing{ID: rpID, RoomTypeID: rtID, RateType: "daily", Rate: 200}

	newRate := 350.0
	resp := doJSON(t, app, "PUT", "/api/admin/room-pricing/"+rpID.String(), map[string]interface{}{
		"rate": newRate,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["rate"] != 350.0 {
		t.Fatalf("expected rate 350, got %v", data["rate"])
	}
}

func TestAdminHandler_UpdateRoomPricing_notFound(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/room-pricing/"+uuid.New().String(), map[string]interface{}{
		"rate": 100,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteRoomPricing_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rpID := uuid.New()
	repos.priceRepo.prices[rpID] = &models.RoomPricing{ID: rpID, RateType: "daily", Rate: 200}

	resp := doJSON(t, app, "DELETE", "/api/admin/room-pricing/"+rpID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteRoomPricing_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "DELETE", "/api/admin/room-pricing/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Inventory tests
// ---------------------------------------------------------------------------

func TestAdminHandler_GetInventory_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	repos.invRepo.records[invKey(rtID, date)] = &models.RoomTypeInventory{
		RoomTypeID: rtID, Date: date, TotalRooms: 10, BookedRooms: 3,
	}

	resp := doJSON(t, app, "GET", fmt.Sprintf("/api/admin/inventory?date=2026-08-01&room_type_id=%s", rtID.String()), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetInventory_missingDate(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/inventory", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetInventory_missingRoomTypeID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/inventory?date=2026-08-01", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetInventory_invalidDate(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/inventory?date=bad&room_type_id="+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateInventory_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	rtID := uuid.New()

	resp := doJSON(t, app, "PUT", "/api/admin/inventory", map[string]interface{}{
		"room_type_id": rtID.String(),
		"date":         "2026-08-01",
		"total_rooms":  10,
		"booked_rooms": 3,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	key := invKey(rtID, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	if _, ok := repos.invRepo.records[key]; !ok {
		t.Fatal("expected inventory to be set")
	}
}

func TestAdminHandler_UpdateInventory_invalidBody(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/inventory", map[string]interface{}{
		"room_type_id": "bad",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Reservation tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListReservations_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reservations", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_ListReservations_invalidFrom(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reservations?from=bad-date", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetReservation_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	resID := uuid.New()
	repos.resRepo.reservations[resID] = &models.Reservation{
		ID: resID, ReferenceCode: "HB-TEST", GuestName: "Alice", Status: "confirmed",
	}

	resp := doJSON(t, app, "GET", "/api/admin/reservations/"+resID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetReservation_notFound(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reservations/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetReservation_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reservations/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateReservationStatus_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	resID := uuid.New()
	repos.resRepo.reservations[resID] = &models.Reservation{
		ID: resID, ReferenceCode: "HB-TEST", Status: "pending",
	}

	resp := doJSON(t, app, "PUT", "/api/admin/reservations/"+resID.String()+"/status", map[string]interface{}{
		"status": "confirmed",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if repos.resRepo.reservations[resID].Status != "confirmed" {
		t.Fatalf("expected confirmed, got %s", repos.resRepo.reservations[resID].Status)
	}
}

func TestAdminHandler_UpdateReservationStatus_invalidStatus(t *testing.T) {
	app, repos := setupAdminHandler(t)
	resID := uuid.New()
	repos.resRepo.reservations[resID] = &models.Reservation{ID: resID, Status: "pending"}

	resp := doJSON(t, app, "PUT", "/api/admin/reservations/"+resID.String()+"/status", map[string]interface{}{
		"status": "bogus",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateReservationStatus_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/reservations/not-a-uuid/status", map[string]interface{}{
		"status": "confirmed",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Customer tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListCustomers_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_ListCustomers_withData(t *testing.T) {
	app, repos := setupAdminHandler(t)
	custID := uuid.New()
	repos.custRepo.customers[custID] = &models.Customer{ID: custID, FullName: "Alice", Email: "alice@test.com"}

	resp := doJSON(t, app, "GET", "/api/admin/customers", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].([]interface{})
	if len(data) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(data))
	}
}

func TestAdminHandler_GetCustomer_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	custID := uuid.New()
	repos.custRepo.customers[custID] = &models.Customer{ID: custID, FullName: "Bob", Email: "bob@test.com"}

	resp := doJSON(t, app, "GET", "/api/admin/customers/"+custID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetCustomer_notFound(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/customers/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_GetCustomer_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/customers/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Admin CRUD tests
// ---------------------------------------------------------------------------

func TestAdminHandler_ListAdmins_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/admins", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateAdmin_ok(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/admins", map[string]interface{}{
		"full_name": "Admin User",
		"email":     "admin@test.com",
		"password":  "secret123",
		"role":      "manager",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["full_name"] != "Admin User" {
		t.Fatalf("expected Admin User, got %v", data["full_name"])
	}
	if data["role"] != "manager" {
		t.Fatalf("expected manager, got %v", data["role"])
	}
}

func TestAdminHandler_CreateAdmin_missingFields(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/admins", map[string]interface{}{
		"email": "admin@test.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_CreateAdmin_defaultRole(t *testing.T) {
	app, repos := setupAdminHandler(t)
	resp := doJSON(t, app, "POST", "/api/admin/admins", map[string]interface{}{
		"full_name": "Front Desk",
		"email":     "fd@test.com",
		"password":  "pass123",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	// find the created admin
	for _, a := range repos.adminRepo.admins {
		if a.Email == "fd@test.com" && a.Role != "front_desk" {
			t.Fatalf("expected default role front_desk, got %s", a.Role)
		}
	}
}

func TestAdminHandler_UpdateAdmin_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	adminID := uuid.New()
	repos.adminRepo.admins[adminID] = &models.Admin{
		ID: adminID, FullName: "Old Name", Email: "old@test.com", Role: "front_desk", IsActive: true,
	}

	newName := "New Name"
	resp := doJSON(t, app, "PUT", "/api/admin/admins/"+adminID.String(), map[string]interface{}{
		"full_name": newName,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if repos.adminRepo.admins[adminID].FullName != "New Name" {
		t.Fatalf("expected New Name, got %s", repos.adminRepo.admins[adminID].FullName)
	}
}

func TestAdminHandler_UpdateAdmin_notFound(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/admins/"+uuid.New().String(), map[string]interface{}{
		"full_name": "x",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_UpdateAdmin_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "PUT", "/api/admin/admins/not-a-uuid", map[string]interface{}{
		"full_name": "x",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteAdmin_ok(t *testing.T) {
	app, repos := setupAdminHandler(t)
	adminID := uuid.New()
	repos.adminRepo.admins[adminID] = &models.Admin{ID: adminID, FullName: "ToDelete"}

	resp := doJSON(t, app, "DELETE", "/api/admin/admins/"+adminID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_DeleteAdmin_invalidID(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "DELETE", "/api/admin/admins/not-a-uuid", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Report tests
// ---------------------------------------------------------------------------

func TestAdminHandler_BookingReport_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/bookings?from=2026-08-01&to=2026-08-31", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["total_bookings"].(float64) != 0 {
		t.Fatalf("expected 0 bookings, got %v", data["total_bookings"])
	}
}

func TestAdminHandler_BookingReport_withData(t *testing.T) {
	app, repos := setupAdminHandler(t)
	repos.resRepo.reservations[uuid.New()] = &models.Reservation{
		GuestName: "A", Status: "confirmed", TotalAmount: 500,
	}
	repos.resRepo.reservations[uuid.New()] = &models.Reservation{
		GuestName: "B", Status: "cancelled", TotalAmount: 300,
	}

	resp := doJSON(t, app, "GET", "/api/admin/reports/bookings?from=2026-01-01&to=2026-12-31", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["total_bookings"].(float64) != 2 {
		t.Fatalf("expected 2 bookings, got %v", data["total_bookings"])
	}
	if data["total_revenue"].(float64) != 500 {
		t.Fatalf("expected 500 revenue, got %v", data["total_revenue"])
	}
}

func TestAdminHandler_BookingReport_invalidFrom(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/bookings?from=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_BookingReport_invalidTo(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/bookings?to=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_OccupancyReport_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/occupancy?from=2026-08-01&to=2026-08-31", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["total_rooms"].(float64) != 29 {
		t.Fatalf("expected 29 total rooms, got %v", data["total_rooms"])
	}
}

func TestAdminHandler_OccupancyReport_invalidFrom(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/occupancy?from=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_RevenueReport_empty(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/revenue?from=2026-08-01&to=2026-08-31", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["total_revenue"].(float64) != 0 {
		t.Fatalf("expected 0 revenue, got %v", data["total_revenue"])
	}
}

func TestAdminHandler_RevenueReport_withData(t *testing.T) {
	app, repos := setupAdminHandler(t)
	repos.resRepo.reservations[uuid.New()] = &models.Reservation{
		GuestName: "A", Status: "confirmed", TotalAmount: 500,
	}
	repos.resRepo.reservations[uuid.New()] = &models.Reservation{
		GuestName: "B", Status: "checked_out", TotalAmount: 800,
	}
	repos.resRepo.reservations[uuid.New()] = &models.Reservation{
		GuestName: "C", Status: "cancelled", TotalAmount: 300,
	}

	resp := doJSON(t, app, "GET", "/api/admin/reports/revenue?from=2026-01-01&to=2026-12-31", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	totalRevenue := data["total_revenue"].(float64)
	if totalRevenue != 1300 { // 500 + 800
		t.Fatalf("expected 1300 revenue, got %v", totalRevenue)
	}
	cancelledRevenue := data["cancelled_revenue"].(float64)
	if cancelledRevenue != 300 {
		t.Fatalf("expected 300 cancelled revenue, got %v", cancelledRevenue)
	}
}

func TestAdminHandler_RevenueReport_invalidFrom(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/revenue?from=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAdminHandler_RevenueReport_invalidTo(t *testing.T) {
	app, _ := setupAdminHandler(t)
	resp := doJSON(t, app, "GET", "/api/admin/reports/revenue?to=bad", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
