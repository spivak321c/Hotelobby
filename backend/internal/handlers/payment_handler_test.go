package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"hotel_lobby/internal/handlers"
	"hotel_lobby/internal/models"
	"hotel_lobby/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Payment repo mock — satisfies services.PaymentUpdater
// ---------------------------------------------------------------------------

type paymentRepoMock struct {
	payments map[string]*models.Payment // keyed by ProviderReference
}

func newPaymentRepoMock() *paymentRepoMock {
	return &paymentRepoMock{payments: make(map[string]*models.Payment)}
}

func (r *paymentRepoMock) FindByReservationID(_ context.Context, id uuid.UUID) (*models.Payment, error) {
	for _, p := range r.payments {
		if p.ReservationID == id {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *paymentRepoMock) FindByProviderReference(_ context.Context, txRef string) (*models.Payment, error) {
	if p, ok := r.payments[txRef]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}

func (r *paymentRepoMock) Create(_ context.Context, p *models.Payment) error {
	r.payments[p.ProviderReference] = p
	return nil
}

func (r *paymentRepoMock) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	for _, p := range r.payments {
		if p.ID == id {
			p.Status = status
			return nil
		}
	}
	return errors.New("not found")
}

type mockWebhookResRepo struct{}

func (m *mockWebhookResRepo) FindByID(_ context.Context, _ uuid.UUID) (*models.Reservation, error) {
	return nil, nil
}
func (m *mockWebhookResRepo) Update(_ context.Context, _ *models.Reservation) error { return nil }

type mockWebhookBookingRepo struct{}

func (m *mockWebhookBookingRepo) FindByReservationID(_ context.Context, _ uuid.UUID) ([]models.Booking, error) {
	return nil, nil
}
func (m *mockWebhookBookingRepo) Update(_ context.Context, _ *models.Booking) error { return nil }
func (m *mockWebhookBookingRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ string) error { return nil }

// ---------------------------------------------------------------------------
// Payment app factory
// ---------------------------------------------------------------------------

type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "staging.crossmint.com" {
		body := `{"orderId": "cm_mock_123", "phase": "processing"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}
	if req.URL.Host == "api.paystack.co" {
		body := `{"status":true,"data":{"reference":"ps_mock_123"}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}
	return http.DefaultTransport.RoundTrip(req)
}

func newPaymentApp(t *testing.T) *fiber.App {
	t.Helper()
	http.DefaultTransport = &mockTransport{}
	svc := services.NewPaymentService("dummy_ps_secret", "dummy_ps_public", "dummy_ps_wh", "dummy_cm_key", "dummy_cm_proj", "dummy_cm_wh", newPaymentRepoMock(), &mockWebhookResRepo{}, &mockWebhookBookingRepo{}, nil)
	h := handlers.NewPaymentHandler(svc)

	app := fiber.New()
	app.Post("/api/payments", h.ProcessPayment)
	app.Get("/api/payments/:reference", h.CheckPayment)
	return app
}

// ---------------------------------------------------------------------------
// Payment handler tests
// ---------------------------------------------------------------------------

func TestPaymentHandler_ProcessPayment_card(t *testing.T) {
	app := newPaymentApp(t)
	resp := doJSON(t, app, "POST", "/api/payments", map[string]string{
		"reservation_id": uuid.New().String(),
		"method":         "card",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	if body["success"] != true {
		t.Fatal("expected success=true")
	}
	data := body["data"].(map[string]interface{})
	if data["status"] != "processing" {
		t.Fatalf("expected status processing, got %v", data["status"])
	}
}

func TestPaymentHandler_ProcessPayment_crypto(t *testing.T) {
	app := newPaymentApp(t)
	resp := doJSON(t, app, "POST", "/api/payments", map[string]string{
		"reservation_id": uuid.New().String(),
		"method":         "crypto",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	data := decodeBody(t, resp)["data"].(map[string]interface{})
	if data["status"] != "processing" {
		t.Fatalf("expected processing, got %v", data["status"])
	}
}

func TestPaymentHandler_ProcessPayment_invalidMethod(t *testing.T) {
	app := newPaymentApp(t)
	resp := doJSON(t, app, "POST", "/api/payments", map[string]string{
		"reservation_id": uuid.New().String(),
		"method":         "bitcoin",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPaymentHandler_ProcessPayment_missingMethod(t *testing.T) {
	app := newPaymentApp(t)
	resp := doJSON(t, app, "POST", "/api/payments", map[string]string{
		"reservation_id": uuid.New().String(),
		// method intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPaymentHandler_CheckPayment_notFound(t *testing.T) {
	app := newPaymentApp(t)
	resp, _ := app.Test(httptest.NewRequest("GET", "/api/payments/NONEXISTENT", nil), 5000)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestPaymentHandler_CheckPayment_seeded(t *testing.T) {
	// Seed repo directly so CheckPayment can find it.
	repo := newPaymentRepoMock()
	repo.payments["PS_abc123"] = &models.Payment{
		ID:                uuid.New(),
		ReservationID:     uuid.New(),
		Amount:            300,
		Provider:          "card",
		Status:            "succeeded",
		ProviderReference: "PS_abc123",
	}
	svc := services.NewPaymentService("", "", "", "", "", "", repo, &mockWebhookResRepo{}, &mockWebhookBookingRepo{}, nil)
	h := handlers.NewPaymentHandler(svc)

	app := fiber.New()
	app.Get("/api/payments/:reference", h.CheckPayment)

	resp, _ := app.Test(httptest.NewRequest("GET", "/api/payments/PS_abc123", nil), 5000)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeBody(t, resp)
	data := body["data"].(map[string]interface{})
	if data["status"] != "succeeded" {
		t.Fatalf("expected succeeded, got %v", data["status"])
	}
	if data["provider_reference"] != "PS_abc123" {
		t.Fatalf("expected PS_abc123, got %v", data["provider_reference"])
	}
}
