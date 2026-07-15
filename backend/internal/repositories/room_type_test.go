package repositories

import (
	"context"
	"testing"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

func createTestRoomType(t *testing.T, repo *RoomTypeRepo) *models.RoomType {
	t.Helper()
	rt := &models.RoomType{
		Name:          "Test Suite",
		Description:   "A test room type",
		BaseRateHourly: 50.00,
		BaseRateDaily:  200.00,
	}
	if err := repo.Create(context.Background(), rt); err != nil {
		t.Fatalf("Create: %v", err)
	}
	return rt
}

func TestRoomTypeRepo_Create(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	rt := createTestRoomType(t, repo)

	if rt.ID == [16]byte{} {
		t.Error("expected non-zero UUID after create")
	}
	if rt.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if rt.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}

	// cleanup
	if err := repo.Delete(context.Background(), rt.ID); err != nil {
		t.Fatalf("cleanup Delete: %v", err)
	}
}

func TestRoomTypeRepo_FindByID(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	rt := createTestRoomType(t, repo)
	defer repo.Delete(context.Background(), rt.ID)

	got, err := repo.FindByID(context.Background(), rt.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name != rt.Name {
		t.Errorf("expected name %q, got %q", rt.Name, got.Name)
	}
	if got.BaseRateHourly != rt.BaseRateHourly {
		t.Errorf("expected hourly rate %.2f, got %.2f", rt.BaseRateHourly, got.BaseRateHourly)
	}
	if got.BaseRateDaily != rt.BaseRateDaily {
		t.Errorf("expected daily rate %.2f, got %.2f", rt.BaseRateDaily, got.BaseRateDaily)
	}
}

func TestRoomTypeRepo_FindByID_notFound(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	_, err := repo.FindByID(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestRoomTypeRepo_Update(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	rt := createTestRoomType(t, repo)
	defer repo.Delete(context.Background(), rt.ID)

	rt.Name = "Updated Suite"
	rt.BaseRateDaily = 250.00
	if err := repo.Update(context.Background(), rt); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(context.Background(), rt.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.Name != "Updated Suite" {
		t.Errorf("expected name 'Updated Suite', got %q", got.Name)
	}
	if got.BaseRateDaily != 250.00 {
		t.Errorf("expected daily rate 250.00, got %.2f", got.BaseRateDaily)
	}
}

func TestRoomTypeRepo_Delete(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	rt := createTestRoomType(t, repo)

	if err := repo.Delete(context.Background(), rt.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(context.Background(), rt.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestRoomTypeRepo_FindAll(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomTypeRepo(pool)

	before, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll before: %v", err)
	}

	rt1 := createTestRoomType(t, repo)
	defer repo.Delete(context.Background(), rt1.ID)
	rt2 := createTestRoomType(t, repo)
	defer repo.Delete(context.Background(), rt2.ID)

	after, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll after: %v", err)
	}
	if len(after) != len(before)+2 {
		t.Errorf("expected %d room types, got %d", len(before)+2, len(after))
	}
}
