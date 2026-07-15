// Package handlers_test contains HTTP-level tests for all handler types.
// Shared helpers and mocks live here; domain-specific tests are in
// auth_handler_test.go, room_handler_test.go, payment_handler_test.go,
// and reservation_handler_test.go.
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// doJSON sends a JSON-encoded body to the Fiber test app.
func doJSON(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

// decodeBody reads and JSON-decodes a response body into a map.
func decodeBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal body: %v\nraw: %s", err, b)
	}
	return m
}

// ---------------------------------------------------------------------------
// Shared customer + admin mocks (used by auth tests)
// ---------------------------------------------------------------------------

// mockCustomerRepo satisfies services.CustomerRepository.
type mockCustomerRepo struct {
	byEmail map[string]*models.Customer
	byID    map[uuid.UUID]*models.Customer
}

func newMockCustomerRepo() *mockCustomerRepo {
	return &mockCustomerRepo{
		byEmail: make(map[string]*models.Customer),
		byID:    make(map[uuid.UUID]*models.Customer),
	}
}

func (r *mockCustomerRepo) FindByEmail(_ context.Context, email string) (*models.Customer, error) {
	if c, ok := r.byEmail[email]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (r *mockCustomerRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Customer, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (r *mockCustomerRepo) Create(_ context.Context, c *models.Customer) error {
	if _, exists := r.byEmail[c.Email]; exists {
		return services.ErrEmailTaken
	}
	r.byEmail[c.Email] = c
	r.byID[c.ID] = c
	return nil
}

// mockAdminRepo satisfies services.AdminRepository.
type mockAdminRepo struct {
	byEmail map[string]*models.Admin
	byID    map[uuid.UUID]*models.Admin
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		byEmail: make(map[string]*models.Admin),
		byID:    make(map[uuid.UUID]*models.Admin),
	}
}

func (r *mockAdminRepo) FindByEmail(_ context.Context, email string) (*models.Admin, error) {
	if a, ok := r.byEmail[email]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func (r *mockAdminRepo) FindByID(_ context.Context, id uuid.UUID) (*models.Admin, error) {
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func (r *mockAdminRepo) Create(_ context.Context, a *models.Admin) error {
	r.byEmail[a.Email] = a
	r.byID[a.ID] = a
	return nil
}
