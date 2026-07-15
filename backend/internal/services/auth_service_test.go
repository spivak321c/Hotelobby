package services

import (
	"context"
	"testing"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockCustomerRepo struct {
	customers map[uuid.UUID]*models.Customer
	byEmail   map[string]*models.Customer
}

func (m *mockCustomerRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	c, ok := m.customers[id]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return c, nil
}

func (m *mockCustomerRepo) FindByEmail(ctx context.Context, email string) (*models.Customer, error) {
	c, ok := m.byEmail[email]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return c, nil
}

func (m *mockCustomerRepo) Create(ctx context.Context, c *models.Customer) error {
	if _, exists := m.byEmail[c.Email]; exists {
		return ErrEmailTaken
	}
	c.ID = uuid.New()
	m.customers[c.ID] = c
	m.byEmail[c.Email] = c
	return nil
}

type mockAdminRepo struct {
	admins  map[uuid.UUID]*models.Admin
	byEmail map[string]*models.Admin
}

func (m *mockAdminRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	a, ok := m.admins[id]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return a, nil
}

func (m *mockAdminRepo) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	a, ok := m.byEmail[email]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return a, nil
}

func (m *mockAdminRepo) Create(ctx context.Context, a *models.Admin) error {
	a.ID = uuid.New()
	m.admins[a.ID] = a
	m.byEmail[a.Email] = a
	return nil
}

func TestAuthService_Register(t *testing.T) {
	cr := &mockCustomerRepo{customers: map[uuid.UUID]*models.Customer{}, byEmail: map[string]*models.Customer{}}
	svc := NewAuthService(cr, nil, "test-secret")

	result, err := svc.Register(context.Background(), "Alice", "alice@test.com", "pass123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	user, ok := result.User.(*models.Customer)
	if !ok {
		t.Fatal("expected *models.Customer")
	}
	if user.FullName != "Alice" {
		t.Errorf("expected Alice, got %s", user.FullName)
	}
}

func TestAuthService_Register_duplicateEmail(t *testing.T) {
	cr := &mockCustomerRepo{customers: map[uuid.UUID]*models.Customer{}, byEmail: map[string]*models.Customer{}}
	svc := NewAuthService(cr, nil, "test-secret")

	svc.Register(context.Background(), "Alice", "alice@test.com", "pass123")
	_, err := svc.Register(context.Background(), "Bob", "alice@test.com", "pass456")
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthService_Login_success(t *testing.T) {
	cr := &mockCustomerRepo{customers: map[uuid.UUID]*models.Customer{}, byEmail: map[string]*models.Customer{}}
	svc := NewAuthService(cr, nil, "test-secret")

	svc.Register(context.Background(), "Alice", "alice@test.com", "pass123")
	result, err := svc.Login(context.Background(), "alice@test.com", "pass123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
}

func TestAuthService_Login_wrongPassword(t *testing.T) {
	cr := &mockCustomerRepo{customers: map[uuid.UUID]*models.Customer{}, byEmail: map[string]*models.Customer{}}
	svc := NewAuthService(cr, nil, "test-secret")

	svc.Register(context.Background(), "Alice", "alice@test.com", "pass123")
	_, err := svc.Login(context.Background(), "alice@test.com", "wrongpass")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_unknownEmail(t *testing.T) {
	cr := &mockCustomerRepo{customers: map[uuid.UUID]*models.Customer{}, byEmail: map[string]*models.Customer{}}
	svc := NewAuthService(cr, nil, "test-secret")

	_, err := svc.Login(context.Background(), "nobody@test.com", "pass123")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_AdminLogin(t *testing.T) {
	pwHash := hashPassword(t, "admin123")
	ar := &mockAdminRepo{admins: map[uuid.UUID]*models.Admin{}, byEmail: map[string]*models.Admin{}}
	ar.Create(context.Background(), &models.Admin{FullName: "Admin", Email: "admin@test.com", PasswordHash: pwHash, Role: "super_admin", IsActive: true})

	svc := NewAuthService(nil, ar, "test-secret")
	result, err := svc.AdminLogin(context.Background(), "admin@test.com", "admin123")
	if err != nil {
		t.Fatalf("AdminLogin: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
}

func hashPassword(t *testing.T, pwd string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(hash)
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc := NewAuthService(nil, nil, "test-secret")

	token, err := svc.generateJWT("user-1", "customer", AccessTokenTTL)
	if err != nil {
		t.Fatalf("generateJWT: %v", err)
	}

	userID, role, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("expected user-1, got %s", userID)
	}
	if role != "customer" {
		t.Errorf("expected customer, got %s", role)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc := NewAuthService(nil, nil, "test-secret")

	result, err := svc.generateTokens("user-1", "customer")
	if err != nil {
		t.Fatalf("generateTokens: %v", err)
	}

	refreshResult, err := svc.RefreshToken(context.Background(), result.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if refreshResult.AccessToken == "" {
		t.Fatal("expected new access token")
	}
	if refreshResult.RefreshToken == "" {
		t.Fatal("expected new refresh token")
	}
}

func TestAuthService_RefreshToken_invalid(t *testing.T) {
	svc := NewAuthService(nil, nil, "test-secret")

	_, err := svc.RefreshToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
}

func TestAuthService_ValidateToken_invalid(t *testing.T) {
	svc := NewAuthService(nil, nil, "test-secret")

	_, _, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
