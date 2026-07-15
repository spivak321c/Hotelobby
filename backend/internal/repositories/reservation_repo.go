package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReservationRepo struct {
	db *pgxpool.Pool
}

func NewReservationRepo(db *pgxpool.Pool) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Reservation, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, reference_code, customer_id, guest_name, guest_email, guest_phone,
		        total_amount, currency, status, cancellation_reason, idempotency_key, created_by_admin_id, created_at, updated_at
		 FROM reservation WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Reservation])
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) FindByReferenceCode(ctx context.Context, referenceCode string) (*models.Reservation, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, reference_code, customer_id, guest_name, guest_email, guest_phone,
		        total_amount, currency, status, cancellation_reason, idempotency_key, created_by_admin_id, created_at, updated_at
		 FROM reservation WHERE reference_code = $1`, referenceCode)
	if err != nil {
		return nil, err
	}
	res, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Reservation])
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]models.Reservation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, reference_code, customer_id, guest_name, guest_email, guest_phone,
		        total_amount, currency, status, cancellation_reason, idempotency_key, created_by_admin_id, created_at, updated_at
		 FROM reservation WHERE customer_id = $1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Reservation])
}

func (r *ReservationRepo) FindAll(ctx context.Context, status string, from, to time.Time) ([]models.Reservation, error) {
	q := `SELECT id, reference_code, customer_id, guest_name, guest_email, guest_phone,
	         total_amount, currency, status, cancellation_reason, idempotency_key, created_by_admin_id, created_at, updated_at
		  FROM reservation WHERE 1=1`
	args := pgx.NamedArgs{}
	if status != "" {
		q += ` AND status = @status`
		args["status"] = status
	}
	if !from.IsZero() {
		q += ` AND created_at >= @from`
		args["from"] = from
	}
	if !to.IsZero() {
		q += ` AND created_at <= @to`
		args["to"] = to
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Reservation])
}

func (r *ReservationRepo) Create(ctx context.Context, res *models.Reservation) error {
	res.ID = uuid.New()
	return r.db.QueryRow(ctx,
		`INSERT INTO reservation (id, reference_code, customer_id, guest_name, guest_email, guest_phone,
		        total_amount, currency, status, idempotency_key, created_by_admin_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING created_at, updated_at`,
		res.ID, res.ReferenceCode, res.CustomerID, res.GuestName, res.GuestEmail,
		res.GuestPhone, res.TotalAmount, res.Currency, res.Status,
		res.IdempotencyKey, res.CreatedByAdminID,
	).Scan(&res.CreatedAt, &res.UpdatedAt)
}

func (r *ReservationRepo) FindByIdempotencyKey(ctx context.Context, key string) (*models.Reservation, error) {
	row, _ := r.db.Query(ctx,
		`SELECT id, reference_code, customer_id, guest_name, guest_email, guest_phone,
		        total_amount, currency, status, cancellation_reason, idempotency_key, created_by_admin_id, created_at, updated_at
		 FROM reservation WHERE idempotency_key = $1`, key)
	res, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Reservation])
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) Update(ctx context.Context, res *models.Reservation) error {
	_, err := r.db.Exec(ctx,
		`UPDATE reservation SET guest_name=$1, guest_email=$2, guest_phone=$3,
		        total_amount=$4, currency=$5, status=$6, idempotency_key=$7,
		        created_by_admin_id=$8, cancellation_reason=$9, updated_at=NOW()
		 WHERE id=$10`,
		res.GuestName, res.GuestEmail, res.GuestPhone,
		res.TotalAmount, res.Currency, res.Status, res.IdempotencyKey,
		res.CreatedByAdminID, res.CancellationReason, res.ID)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *ReservationRepo) FindBookingsByReservationID(ctx context.Context, reservationID uuid.UUID) ([]models.Booking, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, amount, status
		 FROM booking WHERE reservation_id = $1`, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Booking])
}

func translateConstraintErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01":
			return ErrRoomUnavailable
		case "23505":
			return ErrDuplicatePayment
		case "23514":
			return ErrConstraintViolation
		}
	}
	msg := err.Error()
	if strings.Contains(msg, "exclusion") || strings.Contains(msg, "conflict") {
		return ErrRoomUnavailable
	}
	return err
}