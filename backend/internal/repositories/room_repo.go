package repositories

import (
	"context"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo struct {
	db *pgxpool.Pool
}

func NewRoomRepo(db *pgxpool.Pool) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) FindAll(ctx context.Context, roomTypeID *uuid.UUID, status string) ([]models.Room, error) {
	q := `SELECT id, room_type_id, room_number, status, created_at, updated_at
		  FROM room WHERE 1=1`
	args := pgx.NamedArgs{}
	if roomTypeID != nil {
		q += ` AND room_type_id = @room_type_id`
		args["room_type_id"] = *roomTypeID
	}
	if status != "" {
		q += ` AND status = @status`
		args["status"] = status
	}
	q += ` ORDER BY room_number`
	rows, err := r.db.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Room])
}

func (r *RoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, room_type_id, room_number, status, created_at, updated_at
		 FROM room WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	room, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Room])
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepo) Create(ctx context.Context, room *models.Room) error {
	room.ID = uuid.New()
	return r.db.QueryRow(ctx,
		`INSERT INTO room (id, room_type_id, room_number, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		room.ID, room.RoomTypeID, room.RoomNumber, room.Status,
	).Scan(&room.CreatedAt, &room.UpdatedAt)
}

func (r *RoomRepo) Update(ctx context.Context, room *models.Room) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room SET room_type_id=$1, room_number=$2, status=$3, updated_at=NOW()
		 WHERE id=$4`,
		room.RoomTypeID, room.RoomNumber, room.Status, room.ID)
	return err
}

func (r *RoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM room WHERE id = $1`, id)
	return err
}

func (r *RoomRepo) CountActiveBookings(ctx context.Context, roomID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM booking
		 WHERE room_id = $1
		   AND status IN ('pending','confirmed','checked_in')
		   AND ends_at > NOW()`, roomID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RoomRepo) CountRooms(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM room`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RoomRepo) IsAvailable(ctx context.Context, roomID uuid.UUID, checkIn, checkOut time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM room r
			WHERE r.id = $1 AND r.status = 'active'
			  AND NOT EXISTS (
				SELECT 1 FROM booking b
				WHERE b.room_id = r.id
				  AND b.status IN ('pending','confirmed','checked_in')
				  AND tstzrange(b.starts_at, b.ends_at) && tstzrange($2, $3)
			  )
		)`, roomID, checkIn, checkOut).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}