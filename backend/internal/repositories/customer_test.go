package repositories

import (
	"context"
	"testing"

	"hotel_lobby/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func createTestCustomer(t *testing.T, repo *CustomerRepo) *models.Customer {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	c := &models.Customer{
		FullName:     "John Doe",
		Email:        "john@example.com",
		Phone:        "+2348012345678",
		PasswordHash: string(hash),
	}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Customer Create: %v", err)
	}
	return c
}

func TestCustomerRepo_Create(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	c := createTestCustomer(t, repo)
	defer repo.Delete(context.Background(), c.ID)

	if c.ID == [16]byte{} {
		t.Error("expected non-zero UUID")
	}
	if c.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestCustomerRepo_FindByID(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	c := createTestCustomer(t, repo)
	defer repo.Delete(context.Background(), c.ID)

	got, err := repo.FindByID(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.FullName != c.FullName {
		t.Errorf("expected name %q, got %q", c.FullName, got.FullName)
	}
	if got.Email != c.Email {
		t.Errorf("expected email %q, got %q", c.Email, got.Email)
	}
	if got.PasswordHash == "" {
		t.Error("expected password_hash to be returned")
	}
}

func TestCustomerRepo_FindByEmail(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	c := createTestCustomer(t, repo)
	defer repo.Delete(context.Background(), c.ID)

	got, err := repo.FindByEmail(context.Background(), c.Email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("expected ID %v, got %v", c.ID, got.ID)
	}
}

func TestCustomerRepo_FindByEmail_notFound(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	_, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent email")
	}
}

func TestCustomerRepo_Update(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	c := createTestCustomer(t, repo)
	defer repo.Delete(context.Background(), c.ID)

	c.FullName = "Jane Doe"
	c.Phone = "+2348098765432"
	if err := repo.Update(context.Background(), c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.FullName != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %q", got.FullName)
	}
	if got.Phone != "+2348098765432" {
		t.Errorf("expected phone '+2348098765432', got %q", got.Phone)
	}
}

func TestCustomerRepo_duplicateEmail(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepo(pool)

	c1 := createTestCustomer(t, repo)
	defer repo.Delete(context.Background(), c1.ID)

	hash, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	c2 := &models.Customer{
		FullName:     "Jane Doe",
		Email:        c1.Email,
		PasswordHash: string(hash),
	}
	err := repo.Create(context.Background(), c2)
	if err == nil {
		repo.Delete(context.Background(), c2.ID)
		t.Fatal("expected error for duplicate email")
	}
}
