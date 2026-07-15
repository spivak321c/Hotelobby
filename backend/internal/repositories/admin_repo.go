package repositories

import (
	"context"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, full_name, email, password_hash, role, is_active, created_at, updated_at
		 FROM admin WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	a, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Admin])
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, full_name, email, password_hash, role, is_active, created_at, updated_at
		 FROM admin WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	a, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.Admin])
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) FindAll(ctx context.Context) ([]models.Admin, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, full_name, email, password_hash, role, is_active, created_at, updated_at
		 FROM admin ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Admin])
}

func (r *AdminRepo) Create(ctx context.Context, a *models.Admin) error {
	a.ID = uuid.New()
	return r.db.QueryRow(ctx,
		`INSERT INTO admin (id, full_name, email, password_hash, role, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		a.ID, a.FullName, a.Email, a.PasswordHash, a.Role, a.IsActive,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (r *AdminRepo) Update(ctx context.Context, a *models.Admin) error {
	_, err := r.db.Exec(ctx,
		`UPDATE admin SET full_name=$1, email=$2, password_hash=$3, role=$4, is_active=$5, updated_at=NOW()
		 WHERE id=$6`,
		a.FullName, a.Email, a.PasswordHash, a.Role, a.IsActive, a.ID)
	return err
}

func (r *AdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM admin WHERE id = $1`, id)
	return err
}