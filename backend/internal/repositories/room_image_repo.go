package repositories

import (
	"context"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomImageRepo struct {
	db *pgxpool.Pool
}

func NewRoomImageRepo(db *pgxpool.Pool) *RoomImageRepo {
	return &RoomImageRepo{db: db}
}

func (r *RoomImageRepo) FindByRoomID(ctx context.Context, roomID uuid.UUID) ([]models.RoomImage, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, room_id, url, is_primary, sort_order
		 FROM room_image WHERE room_id = $1 ORDER BY sort_order`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.RoomImage])
}

func (r *RoomImageRepo) Create(ctx context.Context, img *models.RoomImage) error {
	img.ID = uuid.New()
	_, err := r.db.Exec(ctx,
		`INSERT INTO room_image (id, room_id, url, is_primary, sort_order)
		 VALUES ($1, $2, $3, $4, $5)`,
		img.ID, img.RoomID, img.URL, img.IsPrimary, img.SortOrder)
	return err
}

func (r *RoomImageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM room_image WHERE id = $1`, id)
	return err
}

func (r *RoomImageRepo) SetPrimary(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_image SET is_primary = (id = $1)
		 WHERE room_id = (SELECT room_id FROM room_image WHERE id = $1)`, id)
	return err
}

func (r *RoomImageRepo) Reorder(ctx context.Context, roomID uuid.UUID, ids []uuid.UUID) error {
	batch := &pgx.Batch{}
	for i, id := range ids {
		batch.Queue(`UPDATE room_image SET sort_order = $1 WHERE id = $2 AND room_id = $3`, i, id, roomID)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range ids {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}