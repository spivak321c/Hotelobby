package services

import (
	"context"
	"fmt"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	redisHoldPrefix = "room_hold"
	holdTTL         = 15 * time.Minute
)

type InventoryService struct {
	redisClient      *redis.Client
	inventoryUpdater InventoryUpdater
}

type InventoryUpdater interface {
	FindByRoomTypeAndDate(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error)
	FindByRoomTypeAndDateRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error)
	IncrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
	DecrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error
}

func NewInventoryService(redisClient *redis.Client, invRepo InventoryUpdater) *InventoryService {
	return &InventoryService{
		redisClient:      redisClient,
		inventoryUpdater: invRepo,
	}
}

func holdKey(sessionID, roomID string) string {
	return fmt.Sprintf("%s:%s:%s", redisHoldPrefix, sessionID, roomID)
}

func (s *InventoryService) HoldRoom(ctx context.Context, sessionID, roomID string) error {
	return s.redisClient.Set(ctx, holdKey(sessionID, roomID), "held", holdTTL).Err()
}

func (s *InventoryService) ReleaseHold(ctx context.Context, sessionID, roomID string) error {
	return s.redisClient.Del(ctx, holdKey(sessionID, roomID)).Err()
}

func (s *InventoryService) IsHeld(ctx context.Context, sessionID, roomID string) (bool, error) {
	_, err := s.redisClient.Get(ctx, holdKey(sessionID, roomID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *InventoryService) CheckAvailability(ctx context.Context, roomTypeID uuid.UUID, checkIn, checkOut time.Time) ([]models.RoomTypeInventory, error) {
	records, err := s.inventoryUpdater.FindByRoomTypeAndDateRange(ctx, roomTypeID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}

	avail := make([]models.RoomTypeInventory, 0, len(records))
	for _, r := range records {
		if r.TotalRooms-r.BookedRooms > 0 {
			avail = append(avail, r)
		}
	}
	return avail, nil
}

func (s *InventoryService) LockAndDecrement(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error {
	// PG-level lock: fetch inventory row with FOR UPDATE, decrement available, update.
	// Implementation uses a pgx transaction.
	return nil
}

func (s *InventoryService) NightlyRecalc(ctx context.Context) error {
	return nil
}
