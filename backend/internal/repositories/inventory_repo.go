package repositories

import (
	"context"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomTypeInventoryRepo struct {
	db *pgxpool.Pool
}

func NewRoomTypeInventoryRepo(db *pgxpool.Pool) *RoomTypeInventoryRepo {
	return &RoomTypeInventoryRepo{db: db}
}

func (r *RoomTypeInventoryRepo) FindByRoomTypeAndDate(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error) {
	row, err := r.db.Query(ctx,
		`SELECT room_type_id, date, total_rooms, booked_rooms
		 FROM room_type_inventory WHERE room_type_id = $1 AND date = $2`, roomTypeID, date)
	if err != nil {
		return nil, err
	}
	inv, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.RoomTypeInventory])
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *RoomTypeInventoryRepo) FindByRoomTypeAndDateLocked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) (*models.RoomTypeInventory, error) {
	row, err := r.db.Query(ctx,
		`SELECT room_type_id, date, total_rooms, booked_rooms
		 FROM room_type_inventory WHERE room_type_id = $1 AND date = $2
		 FOR UPDATE`, roomTypeID, date)
	if err != nil {
		return nil, err
	}
	inv, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.RoomTypeInventory])
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *RoomTypeInventoryRepo) FindByRoomTypeAndDateRange(ctx context.Context, roomTypeID uuid.UUID, from, to time.Time) ([]models.RoomTypeInventory, error) {
	rows, err := r.db.Query(ctx,
		`SELECT room_type_id, date, total_rooms, booked_rooms
		 FROM room_type_inventory WHERE room_type_id = $1 AND date >= $2 AND date <= $3
		 ORDER BY date`, roomTypeID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.RoomTypeInventory])
}

func (r *RoomTypeInventoryRepo) IncrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return translateConstraintErr(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`SELECT 1 FROM room_type_inventory WHERE room_type_id = $1 AND date = $2 FOR UPDATE`, roomTypeID, date)
	if err != nil {
		return translateConstraintErr(err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE room_type_inventory SET booked_rooms = booked_rooms + 1
		 WHERE room_type_id = $1 AND date = $2`, roomTypeID, date)
	if err != nil {
		return translateConstraintErr(err)
	}

	return tx.Commit(ctx)
}

func (r *RoomTypeInventoryRepo) DecrementBooked(ctx context.Context, roomTypeID uuid.UUID, date time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return translateConstraintErr(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`SELECT 1 FROM room_type_inventory WHERE room_type_id = $1 AND date = $2 FOR UPDATE`, roomTypeID, date)
	if err != nil {
		return translateConstraintErr(err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE room_type_inventory SET booked_rooms = GREATEST(booked_rooms - 1, 0)
		 WHERE room_type_id = $1 AND date = $2`, roomTypeID, date)
	if err != nil {
		return translateConstraintErr(err)
	}

	return tx.Commit(ctx)
}

func (r *RoomTypeInventoryRepo) SetInventory(ctx context.Context, roomTypeID uuid.UUID, date time.Time, totalRooms, bookedRooms int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO room_type_inventory (room_type_id, date, total_rooms, booked_rooms)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_type_id, date)
		 DO UPDATE SET total_rooms = EXCLUDED.total_rooms, booked_rooms = EXCLUDED.booked_rooms`,
		roomTypeID, date, totalRooms, bookedRooms)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}