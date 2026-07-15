package handlers_test

import (
	"net/http"
	"testing"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
)

// newAuthApp wires up a minimal Fiber app containing only the auth routes.
func newAuthApp(t *testing.T) (*fiber.App, *mockCustomerRepo) {
	t.Helper()
	cr := newMockCustomerRepo()
	ar := newMockAdminRepo()
	authSvc := services.NewAuthService(cr, ar, "test-secret-key")
	h := handlers.NewAuthHandler(authSvc)

	app := fiber.New()
	app.Post("/api/auth/register", h.Register)
	app.Post("/api/auth/login", h.Login)
	app.Post("/api/auth/admin/login", h.AdminLogin)
	return app, cr
}

func TestAuthHandler_Register_ok(t *testing.T) {
	app, _ := newAuthApp(t)
	resp := doJSON(t, app, "POST", "/api/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "secret123",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if _, ok := decodeBody(t, resp)["access_token"]; !ok {
		t.Fatal("expected access_token in response")
	}
}

func TestAuthHandler_Register_missingFields(t *testing.T) {
	app, _ := newAuthApp(t)
	resp := doJSON(t, app, "POST", "/api/auth/register", map[string]string{
		"email": "alice@example.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Register_duplicateEmail(t *testing.T) {
	app, _ := newAuthApp(t)
	payload := map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "secret123",
	}
	doJSON(t, app, "POST", "/api/auth/register", payload)
	resp := doJSON(t, app, "POST", "/api/auth/register", payload)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_Login_ok(t *testing.T) {
	app, _ := newAuthApp(t)
	doJSON(t, app, "POST", "/api/auth/register", map[string]string{
		"name": "Bob", "email": "bob@example.com", "password": "mypassword",
	})
	resp := doJSON(t, app, "POST", "/api/auth/login", map[string]string{
		"email": "bob@example.com", "password": "mypassword",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if _, ok := decodeBody(t, resp)["access_token"]; !ok {
		t.Fatal("expected access_token in response")
	}
}

func TestAuthHandler_Login_wrongPassword(t *testing.T) {
	app, _ := newAuthApp(t)
	doJSON(t, app, "POST", "/api/auth/register", map[string]string{
		"name": "Carol", "email": "carol@example.com", "password": "correct",
	})
	resp := doJSON(t, app, "POST", "/api/auth/login", map[string]string{
		"email": "carol@example.com", "password": "wrong",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthHandler_AdminLogin_notFound(t *testing.T) {
	app, _ := newAuthApp(t)
	resp := doJSON(t, app, "POST", "/api/auth/admin/login", map[string]string{
		"email": "admin@example.com", "password": "adminpass",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
