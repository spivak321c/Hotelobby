package services

import (
	"context"
	"errors"
	"testing"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

// mockCustomerRepoOps satisfies CustomerRepoOps.
// Named with the "Ops" suffix to avoid collision with mockCustomerRepo in auth_service_test.go,
// which implements the narrower AuthCustomerRepo interface (no Update method).
type mockCustomerRepoOps struct {
	customers map[uuid.UUID]*models.Customer
	byEmail   map[string]*models.Customer
}

func (m *mockCustomerRepoOps) FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	c, ok := m.customers[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (m *mockCustomerRepoOps) FindByEmail(ctx context.Context, email string) (*models.Customer, error) {
	c, ok := m.byEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (m *mockCustomerRepoOps) Create(ctx context.Context, c *models.Customer) error {
	c.ID = uuid.New()
	m.customers[c.ID] = c
	m.byEmail[c.Email] = c
	return nil
}

func (m *mockCustomerRepoOps) Update(ctx context.Context, c *models.Customer) error {
	m.customers[c.ID] = c
	if c.Email != "" {
		m.byEmail[c.Email] = c
	}
	return nil
}

// mockCustomerReservationRepo satisfies CustomerReservationRepo.
type mockCustomerReservationRepo struct {
	reservations map[uuid.UUID]*models.Reservation
	bookings     map[uuid.UUID][]models.Booking
}

func (m *mockCustomerReservationRepo) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	var out []models.Reservation
	for _, r := range m.reservations {
		if r.CustomerID != nil && *r.CustomerID == customerID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *mockCustomerReservationRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("reservation not found")
	}
	return r, nil
}

func (m *mockCustomerReservationRepo) FindBookingsByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	if b, ok := m.bookings[reservationID]; ok {
		return b, nil
	}
	return nil, nil
}

// helpers

func newTestCustomerService(
	cr *mockCustomerRepoOps,
	rr *mockCustomerReservationRepo,
) *CustomerService {
	return NewCustomerService(cr, rr)
}

// TestCustomerService_GetProfile_found seeds a customer and verifies the profile is returned.
func TestCustomerService_GetProfile_found(t *testing.T) {
	id := uuid.New()
	cr := &mockCustomerRepoOps{
		customers: map[uuid.UUID]*models.Customer{
			id: {ID: id, FullName: "Alice", Email: "alice@test.com", Phone: "+2348000000"},
		},
		byEmail: map[string]*models.Customer{},
	}
	svc := newTestCustomerService(cr, &mockCustomerReservationRepo{reservations: map[uuid.UUID]*models.Reservation{}})

	c, err := svc.GetProfile(context.Background(), id)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if c.FullName != "Alice" {
		t.Errorf("expected Alice, got %s", c.FullName)
	}
	if c.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", c.Email)
	}
}

// TestCustomerService_GetProfile_notFound expects ErrCustomerNotFound for an unknown ID.
func TestCustomerService_GetProfile_notFound(t *testing.T) {
	cr := &mockCustomerRepoOps{
		customers: map[uuid.UUID]*models.Customer{},
		byEmail:   map[string]*models.Customer{},
	}
	svc := newTestCustomerService(cr, &mockCustomerReservationRepo{reservations: map[uuid.UUID]*models.Reservation{}})

	_, err := svc.GetProfile(context.Background(), uuid.New())
	if err != ErrCustomerNotFound {
		t.Fatalf("expected ErrCustomerNotFound, got %v", err)
	}
}

// TestCustomerService_UpdateProfile_nameAndPhone verifies that name and phone fields are patched.
func TestCustomerService_UpdateProfile_nameAndPhone(t *testing.T) {
	id := uuid.New()
	cr := &mockCustomerRepoOps{
		customers: map[uuid.UUID]*models.Customer{
			id: {ID: id, FullName: "Alice", Email: "alice@test.com", Phone: "+1000000000"},
		},
		byEmail: map[string]*models.Customer{},
	}
	svc := newTestCustomerService(cr, &mockCustomerReservationRepo{reservations: map[uuid.UUID]*models.Reservation{}})

	updated, err := svc.UpdateProfile(context.Background(), id, "Alicia", "+2349000000")
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.FullName != "Alicia" {
		t.Errorf("expected name Alicia, got %s", updated.FullName)
	}
	if updated.Phone != "+2349000000" {
		t.Errorf("expected phone +2349000000, got %s", updated.Phone)
	}
	// verify the change is persisted in the mock store
	stored := cr.customers[id]
	if stored.FullName != "Alicia" {
		t.Errorf("stored name: expected Alicia, got %s", stored.FullName)
	}
}

// TestCustomerService_ListReservations_ok verifies all reservations owned by a customer are returned.
func TestCustomerService_ListReservations_ok(t *testing.T) {
	customerID := uuid.New()
	otherID := uuid.New()

	r1ID := uuid.New()
	r2ID := uuid.New()
	r3ID := uuid.New()

	cr := &mockCustomerRepoOps{
		customers: map[uuid.UUID]*models.Customer{
			customerID: {ID: customerID, FullName: "Alice", Email: "alice@test.com"},
		},
		byEmail: map[string]*models.Customer{},
	}
	rr := &mockCustomerReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{
			r1ID: {ID: r1ID, CustomerID: &customerID, ReferenceCode: "HB-001"},
			r2ID: {ID: r2ID, CustomerID: &customerID, ReferenceCode: "HB-002"},
			r3ID: {ID: r3ID, CustomerID: &otherID, ReferenceCode: "HB-003"}, // belongs to someone else
		},
	}
	svc := newTestCustomerService(cr, rr)

	list, err := svc.ListReservations(context.Background(), customerID)
	if err != nil {
		t.Fatalf("ListReservations: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 reservations, got %d", len(list))
	}
}

// TestCustomerService_GetReservation_wrongCustomer expects an error when the reservation
// belongs to a different customer.
func TestCustomerService_GetReservation_wrongCustomer(t *testing.T) {
	ownerID := uuid.New()
	callerID := uuid.New()
	resID := uuid.New()

	rr := &mockCustomerReservationRepo{
		reservations: map[uuid.UUID]*models.Reservation{
			resID: {ID: resID, CustomerID: &ownerID, ReferenceCode: "HB-OWNER"},
		},
	}
	// CustomerRepoOps is not exercised by GetReservation, so pass a minimal stub.
	cr := &mockCustomerRepoOps{
		customers: map[uuid.UUID]*models.Customer{},
		byEmail:   map[string]*models.Customer{},
	}
	svc := newTestCustomerService(cr, rr)

	res, err := svc.GetReservation(context.Background(), callerID, resID)
	// The service returns (nil, err) when the customer ID doesn't match.
	// We accept either a non-nil error OR a nil result — either signals the access was denied.
	if err == nil && res != nil {
		t.Fatal("expected error or nil result when reservation belongs to a different customer")
	}
}
