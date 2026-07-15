package repositories

import (
	"context"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) FindByReservationID(ctx context.Context, reservationID uuid.UUID) (*models.Payment, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, reservation_id, provider, provider_reference, amount, currency, metadata, status, created_at, updated_at
		 FROM payment WHERE reservation_id = $1`, reservationID)
	if err != nil {
		return nil, err
	}
	p, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Payment])
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) FindByProviderReference(ctx context.Context, providerRef string) (*models.Payment, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, reservation_id, provider, provider_reference, amount, currency, metadata, status, created_at, updated_at
		 FROM payment WHERE provider_reference = $1`, providerRef)
	if err != nil {
		return nil, err
	}
	p, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Payment])
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) Create(ctx context.Context, p *models.Payment) error {
	p.ID = uuid.New()
	_, err := r.db.Exec(ctx,
		`INSERT INTO payment (id, reservation_id, provider, provider_reference, amount, currency, metadata, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.ReservationID, p.Provider, p.ProviderReference, p.Amount, p.Currency, p.Metadata, p.Status)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *PaymentRepo) Update(ctx context.Context, p *models.Payment) error {
	_, err := r.db.Exec(ctx,
		`UPDATE payment SET provider=$1, provider_reference=$2, amount=$3, currency=$4,
		        metadata=$5, status=$6, updated_at=NOW()
		 WHERE id=$7`,
		p.Provider, p.ProviderReference, p.Amount, p.Currency, p.Metadata, p.Status, p.ID)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE payment SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}