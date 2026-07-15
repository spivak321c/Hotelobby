package repositories

import (
	"context"
	"fmt"
	"testing"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
)

var roomCounter int

func nextRoomNumber() string {
	roomCounter++
	return fmt.Sprintf("T%d", roomCounter)
}

func createTestRoom(t *testing.T, roomTypeRepo *RoomTypeRepo, roomRepo *RoomRepo) (*models.RoomType, *models.Room) {
	t.Helper()
	rt := createTestRoomType(t, roomTypeRepo)
	t.Cleanup(func() { roomTypeRepo.Delete(context.Background(), rt.ID) })

	r := &models.Room{
		RoomTypeID: rt.ID,
		RoomNumber: nextRoomNumber(),
		Status:     "available",
	}
	if err := roomRepo.Create(context.Background(), r); err != nil {
		t.Fatalf("Room Create: %v", err)
	}
	return rt, r
}

func TestRoomRepo_Create(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	_, r := createTestRoom(t, rtRepo, roomRepo)

	if r.ID == [16]byte{} {
		t.Error("expected non-zero UUID")
	}
	if r.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}

	// cleanup
	if err := roomRepo.Delete(context.Background(), r.ID); err != nil {
		t.Fatalf("cleanup Delete: %v", err)
	}
}

func TestRoomRepo_FindByID(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	_, r := createTestRoom(t, rtRepo, roomRepo)
	defer roomRepo.Delete(context.Background(), r.ID)

	got, err := roomRepo.FindByID(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.RoomNumber != r.RoomNumber {
		t.Errorf("expected room number %q, got %q", r.RoomNumber, got.RoomNumber)
	}
}

func TestRoomRepo_FindByID_notFound(t *testing.T) {
	pool := testPool(t)
	repo := NewRoomRepo(pool)

	_, err := repo.FindByID(context.Background(), uuid.Nil)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestRoomRepo_FindAll_filterByStatus(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	_, r := createTestRoom(t, rtRepo, roomRepo)
	defer roomRepo.Delete(context.Background(), r.ID)

	all, err := roomRepo.FindAll(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) < 1 {
		t.Fatal("expected at least 1 room")
	}

	available, err := roomRepo.FindAll(context.Background(), nil, "available")
	if err != nil {
		t.Fatalf("FindAll by status: %v", err)
	}
	if len(available) < 1 {
		t.Fatal("expected at least 1 available room")
	}

	maintenance, err := roomRepo.FindAll(context.Background(), nil, "maintenance")
	if err != nil {
		t.Fatalf("FindAll by maintenance: %v", err)
	}
	for _, rm := range maintenance {
		if rm.Status != "maintenance" {
			t.Errorf("expected status maintenance, got %q", rm.Status)
		}
	}
}

func TestRoomRepo_Update(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	_, r := createTestRoom(t, rtRepo, roomRepo)
	defer roomRepo.Delete(context.Background(), r.ID)

	r.Status = "maintenance"
	if err := roomRepo.Update(context.Background(), r); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := roomRepo.FindByID(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.Status != "maintenance" {
		t.Errorf("expected status maintenance, got %q", got.Status)
	}
}

func TestRoomRepo_Delete(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	_, r := createTestRoom(t, rtRepo, roomRepo)

	if err := roomRepo.Delete(context.Background(), r.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := roomRepo.FindByID(context.Background(), r.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestRoomRepo_duplicateRoomNumber(t *testing.T) {
	pool := testPool(t)
	roomRepo := NewRoomRepo(pool)
	rtRepo := NewRoomTypeRepo(pool)

	rt, r1 := createTestRoom(t, rtRepo, roomRepo)
	defer roomRepo.Delete(context.Background(), r1.ID)

	r2 := &models.Room{
		RoomTypeID: rt.ID,
		RoomNumber: r1.RoomNumber,
		Status:     "available",
	}
	err := roomRepo.Create(context.Background(), r2)
	if err == nil {
		roomRepo.Delete(context.Background(), r2.ID)
		t.Fatal("expected error for duplicate room number")
	}
}
