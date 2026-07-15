package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/middleware"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Customer service mock
// ---------------------------------------------------------------------------

type custServCustomerRepo struct {
	customers map[uuid.UUID]*models.Customer
	byEmail   map[string]*models.Customer
}

func newCustServCustomerRepo() *custServCustomerRepo {
	return &custServCustomerRepo{
		customers: make(map[uuid.UUID]*models.Customer),
		byEmail:   make(map[string]*models.Customer),
	}
}

func (r *custServCustomerRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Customer, error) {
	if c, ok := r.customers[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}
func (r *custServCustomerRepo) FindByEmail(_ context.Context, email string) (*models.Customer, error) {
	if c, ok := r.byEmail[email]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}
func (r *custServCustomerRepo) Create(_ context.Context, c *models.Customer) error {
	r.customers[c.ID] = c
	r.byEmail[c.Email] = c
	return nil
}
func (r *custServCustomerRepo) Update(_ context.Context, c *models.Customer) error {
	r.customers[c.ID] = c
	return nil
}
func (r *custServCustomerRepo) FindAll(_ context.Context) ([]models.Customer, error) {
	out := make([]models.Customer, 0, len(r.customers))
	for _, c := range r.customers {
		out = append(out, *c)
	}
	return out, nil
}
func (r *custServCustomerRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.customers, id)
	return nil
}

type custServResRepo struct {
	reservations map[uuid.UUID]*models.Reservation
}

func newCustServResRepo() *custServResRepo {
	return &custServResRepo{reservations: make(map[uuid.UUID]*models.Reservation)}
}

func (r *custServResRepo) FindByCustomerID(_ context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	var out []models.Reservation
	for _, res := range r.reservations {
		if res.CustomerID != nil && *res.CustomerID == customerID {
			out = append(out, *res)
		}
	}
	return out, nil
}
func (r *custServResRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Reservation, error) {
	if res, ok := r.reservations[id]; ok {
		return res, nil
	}
	return nil, errors.New("not found")
}
func (r *custServResRepo) FindBookingsByReservationID(_ context.Context, _ uuid.UUID) ([]models.Booking, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// CustomerHandler wiring
// ---------------------------------------------------------------------------

func setupCustomerHandler(t *testing.T) (*fiber.App, *custServCustomerRepo, *custServResRepo) {
	t.Helper()

	custRepo := newCustServCustomerRepo()
	resRepo := newCustServResRepo()
	custSvc := services.NewCustomerService(custRepo, resRepo)
	h := handlers.NewCustomerHandler(custSvc)

	app := fiber.New()
	// Middleware to inject user_id from the X-User-ID header into Fiber locals.
	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals(middleware.KeyUserID, uid)
		}
		return c.Next()
	})
	app.Get("/api/customer/profile", h.GetProfile)
	app.Put("/api/customer/profile", h.UpdateProfile)
	app.Get("/api/customer/reservations", h.ListReservations)
	app.Get("/api/customer/reservations/:id", h.GetReservation)

	return app, custRepo, resRepo
}

// doJSONWithAuth sends a request with user_id injected via the X-User-ID header.
func doJSONWithAuth(t *testing.T, app *fiber.App, method, path string, body interface{}, userID string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

func TestCustomerHandler_GetProfile_ok(t *testing.T) {
	app, custRepo, _ := setupCustomerHandler(t)
	custID := uuid.New()
	custRepo.customers[custID] = &models.Customer{
		ID: custID, FullName: "Alice", Email: "alice@test.com", Phone: "+2348000000",
	}

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/profile", nil, custID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["full_name"] != "Alice" {
		t.Fatalf("expected Alice, got %v", data["full_name"])
	}
}

func TestCustomerHandler_GetProfile_unauthorized(t *testing.T) {
	app, _, _ := setupCustomerHandler(t)
	resp := doJSON(t, app, "GET", "/api/customer/profile", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_GetProfile_notFound(t *testing.T) {
	app, _, _ := setupCustomerHandler(t)
	unknownID := uuid.New()
	resp := doJSONWithAuth(t, app, "GET", "/api/customer/profile", nil, unknownID.String())
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_UpdateProfile_ok(t *testing.T) {
	app, custRepo, _ := setupCustomerHandler(t)
	custID := uuid.New()
	custRepo.customers[custID] = &models.Customer{
		ID: custID, FullName: "Alice", Email: "alice@test.com",
	}

	resp := doJSONWithAuth(t, app, "PUT", "/api/customer/profile", map[string]interface{}{
		"name":  "Alicia",
		"phone": "+2349000000",
	}, custID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_UpdateProfile_unauthorized(t *testing.T) {
	app, _, _ := setupCustomerHandler(t)
	resp := doJSON(t, app, "PUT", "/api/customer/profile", map[string]interface{}{
		"name": "x",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_ListReservations_ok(t *testing.T) {
	app, custRepo, resRepo := setupCustomerHandler(t)
	custID := uuid.New()
	custRepo.customers[custID] = &models.Customer{ID: custID, FullName: "Alice", Email: "alice@test.com"}
	resID1 := uuid.New()
	resID2 := uuid.New()
	resRepo.reservations[resID1] = &models.Reservation{
		ID: resID1, CustomerID: &custID, ReferenceCode: "HB-001", TotalAmount: 400, Status: "confirmed",
	}
	resRepo.reservations[resID2] = &models.Reservation{
		ID: resID2, CustomerID: &custID, ReferenceCode: "HB-002", TotalAmount: 600, Status: "pending",
	}

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations", nil, custID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].([]interface{})
	if len(data) != 2 {
		t.Fatalf("expected 2 reservations, got %d", len(data))
	}
}

func TestCustomerHandler_ListReservations_empty(t *testing.T) {
	app, custRepo, _ := setupCustomerHandler(t)
	custID := uuid.New()
	custRepo.customers[custID] = &models.Customer{ID: custID, FullName: "Alice", Email: "alice@test.com"}

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations", nil, custID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_ListReservations_unknownCustomer(t *testing.T) {
	app, _, _ := setupCustomerHandler(t)
	unknownID := uuid.New()

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations", nil, unknownID.String())
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_GetReservation_ok(t *testing.T) {
	app, _, resRepo := setupCustomerHandler(t)
	custID := uuid.New()
	resID := uuid.New()
	_ = time.Now()
	resRepo.reservations[resID] = &models.Reservation{
		ID: resID, CustomerID: &custID, ReferenceCode: "HB-001",
		GuestName: "Alice", GuestEmail: "alice@test.com", TotalAmount: 400, Status: "confirmed",
	}

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations/"+resID.String(), nil, custID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_GetReservation_wrongCustomer(t *testing.T) {
	app, _, resRepo := setupCustomerHandler(t)
	ownerID := uuid.New()
	callerID := uuid.New()
	resID := uuid.New()
	resRepo.reservations[resID] = &models.Reservation{
		ID: resID, CustomerID: &ownerID, ReferenceCode: "HB-OWN",
	}

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations/"+resID.String(), nil, callerID.String())
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCustomerHandler_GetReservation_invalidID(t *testing.T) {
	app, _, _ := setupCustomerHandler(t)
	custID := uuid.New()

	resp := doJSONWithAuth(t, app, "GET", "/api/customer/reservations/not-a-uuid", nil, custID.String())
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
