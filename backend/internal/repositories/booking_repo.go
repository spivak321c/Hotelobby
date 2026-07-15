package repositories

import (
	"context"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, amount, status
		 FROM booking WHERE reservation_id = $1`, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Booking])
}

func (r *BookingRepo) FindByReservationIDBatch(ctx context.Context, reservationIDs []uuid.UUID) ([]models.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, amount, status
		 FROM booking WHERE reservation_id = ANY($1::uuid[]) ORDER BY starts_at`, reservationIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Booking])
}

func (r *BookingRepo) Create(ctx context.Context, b *models.Booking) error {
	b.ID = uuid.New()
	_, err := r.db.Exec(ctx,
		`INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		b.ID, b.ReservationID, b.RoomID, b.RoomTypeID, b.BookingType, b.StartsAt, b.EndsAt, b.Amount, b.Status)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *BookingRepo) Update(ctx context.Context, b *models.Booking) error {
	_, err := r.db.Exec(ctx,
		`UPDATE booking SET room_id=$1, room_type_id=$2, booking_type=$3,
		        starts_at=$4, ends_at=$5, amount=$6, status=$7
		 WHERE id=$8`,
		b.RoomID, b.RoomTypeID, b.BookingType, b.StartsAt, b.EndsAt, b.Amount, b.Status, b.ID)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE booking SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *BookingRepo) FindAvailableRoom(ctx context.Context, roomTypeID uuid.UUID, startsAt, endsAt time.Time) (uuid.UUID, error) {
	var roomID uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT r.id FROM room r
		 WHERE r.room_type_id = $1 AND r.status = 'active'
		   AND NOT EXISTS (
		       SELECT 1 FROM booking b
		       WHERE b.room_id = r.id
		         AND b.status IN ('pending','confirmed','checked_in')
		         AND tstzrange(b.starts_at, b.ends_at) && tstzrange($2, $3)
		   )
		 LIMIT 1 FOR UPDATE SKIP LOCKED`,
		roomTypeID, startsAt, endsAt,
	).Scan(&roomID)
	if err != nil {
		return uuid.Nil, err
	}
	return roomID, nil
}