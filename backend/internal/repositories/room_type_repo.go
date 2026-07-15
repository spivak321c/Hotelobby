package repositories

import (
	"context"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomTypeRepo struct {
	db *pgxpool.Pool
}

func NewRoomTypeRepo(db *pgxpool.Pool) *RoomTypeRepo {
	return &RoomTypeRepo{db: db}
}

func (r *RoomTypeRepo) FindAll(ctx context.Context) ([]models.RoomType, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, base_rate_daily, base_rate_hourly, max_occupancy, is_featured, created_at, updated_at
		 FROM room_type ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.RoomType])
}

func (r *RoomTypeRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.RoomType, error) {
	row, err := r.db.Query(ctx,
		`SELECT id, name, description, base_rate_daily, base_rate_hourly, max_occupancy, is_featured, created_at, updated_at
		 FROM room_type WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	rt, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[models.RoomType])
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RoomTypeRepo) Create(ctx context.Context, rt *models.RoomType) error {
	rt.ID = uuid.New()
	return r.db.QueryRow(ctx,
		`INSERT INTO room_type (id, name, description, base_rate_daily, base_rate_hourly, max_occupancy, is_featured)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		rt.ID, rt.Name, rt.Description, rt.BaseRateDaily, rt.BaseRateHourly, rt.MaxOccupancy, rt.IsFeatured,
	).Scan(&rt.CreatedAt, &rt.UpdatedAt)
}

func (r *RoomTypeRepo) Update(ctx context.Context, rt *models.RoomType) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_type SET name=$1, description=$2, base_rate_daily=$3, base_rate_hourly=$4,
		        max_occupancy=$5, is_featured=$6, updated_at=NOW()
		 WHERE id=$7`,
		rt.Name, rt.Description, rt.BaseRateDaily, rt.BaseRateHourly, rt.MaxOccupancy, rt.IsFeatured, rt.ID)
	return err
}

func (r *RoomTypeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM room_type WHERE id = $1`, id)
	return err
}