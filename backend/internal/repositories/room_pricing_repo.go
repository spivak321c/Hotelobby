package repositories

import (
	"context"
	"fmt"

	"hotel_lobby/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomPricingRepo struct {
	db *pgxpool.Pool
}

func NewRoomPricingRepo(db *pgxpool.Pool) *RoomPricingRepo {
	return &RoomPricingRepo{db: db}
}

func scanEffectiveRange(src *pgtype.Range[pgtype.Date], dst *models.Daterange) {
	if !src.Valid {
		return
	}
	if src.Lower.Valid {
		dst.Lower = src.Lower.Time
	}
	if src.Upper.Valid {
		dst.Upper = src.Upper.Time
	}
	dst.Bounds = fmt.Sprintf("[%s,%s)",
		dst.Lower.Format("2006-01-02"),
		dst.Upper.Format("2006-01-02"),
	)
}

func (r *RoomPricingRepo) FindAll(ctx context.Context, roomTypeID *uuid.UUID) ([]models.RoomPricing, error) {
	q := `SELECT id, room_type_id, rate_type, rate, effective_range
		  FROM room_pricing WHERE 1=1`
	args := pgx.NamedArgs{}
	if roomTypeID != nil {
		q += ` AND room_type_id = @room_type_id`
		args["room_type_id"] = *roomTypeID
	}
	q += ` ORDER BY lower(effective_range)`
	rows, err := r.db.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.RoomPricing
	for rows.Next() {
		var rp models.RoomPricing
		var dr pgtype.Range[pgtype.Date]
		if err := rows.Scan(&rp.ID, &rp.RoomTypeID, &rp.RateType, &rp.Rate, &dr); err != nil {
			rows.Close()
			return nil, err
		}
		scanEffectiveRange(&dr, &rp.EffectiveRange)
		results = append(results, rp)
	}
	return results, rows.Err()
}

func (r *RoomPricingRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.RoomPricing, error) {
	var rp models.RoomPricing
	var dr pgtype.Range[pgtype.Date]
	row, err := r.db.Query(ctx,
		`SELECT id, room_type_id, rate_type, rate, effective_range
		 FROM room_pricing WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	got, err := pgx.CollectExactlyOneRow(row, func(rows pgx.CollectableRow) (models.RoomPricing, error) {
		if err := rows.Scan(&rp.ID, &rp.RoomTypeID, &rp.RateType, &rp.Rate, &dr); err != nil {
			return rp, err
		}
		scanEffectiveRange(&dr, &rp.EffectiveRange)
		return rp, nil
	})
	if err != nil {
		return nil, err
	}
	return &got, nil
}

func (r *RoomPricingRepo) FindByRoomTypeID(ctx context.Context, roomTypeID uuid.UUID) ([]models.RoomPricing, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, room_type_id, rate_type, rate, effective_range
		 FROM room_pricing WHERE room_type_id = $1 ORDER BY lower(effective_range)`, roomTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.RoomPricing
	for rows.Next() {
		var rp models.RoomPricing
		var dr pgtype.Range[pgtype.Date]
		if err := rows.Scan(&rp.ID, &rp.RoomTypeID, &rp.RateType, &rp.Rate, &dr); err != nil {
			rows.Close()
			return nil, err
		}
		scanEffectiveRange(&dr, &rp.EffectiveRange)
		results = append(results, rp)
	}
	return results, rows.Err()
}

func (r *RoomPricingRepo) Create(ctx context.Context, rp *models.RoomPricing) error {
	rp.ID = uuid.New()
	var lower pgtype.Date
	var upper pgtype.Date
	if !rp.EffectiveRange.Lower.IsZero() {
		lower = pgtype.Date{Time: rp.EffectiveRange.Lower, Valid: true}
	}
	if !rp.EffectiveRange.Upper.IsZero() {
		upper = pgtype.Date{Time: rp.EffectiveRange.Upper, Valid: true}
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO room_pricing (id, room_type_id, rate_type, rate, effective_range)
		 VALUES ($1, $2, $3, $4, daterange($5::date, $6::date, '[]'))`,
		rp.ID, rp.RoomTypeID, rp.RateType, rp.Rate, lower, upper)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *RoomPricingRepo) Update(ctx context.Context, rp *models.RoomPricing) error {
	var lower pgtype.Date
	var upper pgtype.Date
	if !rp.EffectiveRange.Lower.IsZero() {
		lower = pgtype.Date{Time: rp.EffectiveRange.Lower, Valid: true}
	}
	if !rp.EffectiveRange.Upper.IsZero() {
		upper = pgtype.Date{Time: rp.EffectiveRange.Upper, Valid: true}
	}

	_, err := r.db.Exec(ctx,
		`UPDATE room_pricing SET rate_type=$1, rate=$2, effective_range=daterange($3::date, $4::date, '[]')
		 WHERE id=$5`,
		rp.RateType, rp.Rate, lower, upper, rp.ID)
	if err != nil {
		return translateConstraintErr(err)
	}
	return nil
}

func (r *RoomPricingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM room_pricing WHERE id = $1`, id)
	return err
}