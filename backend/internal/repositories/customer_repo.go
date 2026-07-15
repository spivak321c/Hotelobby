package repositories

import (
	"context"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepo struct {
	db *pgxpool.Pool
}

func NewCustomerRepo(db *pgxpool.Pool) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) FindAll(ctx context.Context) ([]models.Customer, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, full_name, email, phone, password_hash, created_at, updated_at
		 FROM customer ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Customer])
}

func (r *CustomerRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, full_name, email, phone, password_hash, created_at, updated_at
		 FROM customer WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	c, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Customer])
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepo) FindByEmail(ctx context.Context, email string) (*models.Customer, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, full_name, email, phone, password_hash, created_at, updated_at
		 FROM customer WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	c, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Customer])
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepo) Create(ctx context.Context, c *models.Customer) error {
	c.ID = uuid.New()
	return r.db.QueryRow(ctx,
		`INSERT INTO customer (id, full_name, email, phone, password_hash)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at, updated_at`,
		c.ID, c.FullName, c.Email, c.Phone, c.PasswordHash,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (r *CustomerRepo) Update(ctx context.Context, c *models.Customer) error {
	_, err := r.db.Exec(ctx,
		`UPDATE customer SET full_name=$1, email=$2, phone=$3, password_hash=$4, updated_at=NOW()
		 WHERE id=$5`,
		c.FullName, c.Email, c.Phone, c.PasswordHash, c.ID)
	return err
}

func (r *CustomerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM customer WHERE id = $1`, id)
	return err
}