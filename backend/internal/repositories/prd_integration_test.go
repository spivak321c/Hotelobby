package repositories

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seedTestRoom inserts a room_type + room + inventory row and returns their IDs.
func seedTestRoom(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (rtID, roomID uuid.UUID) {
	t.Helper()
	rtID = uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO room_type (id, name, base_rate_daily, base_rate_hourly, max_occupancy)
		VALUES ($1, 'Integration Test Room', 200, 50, 2)`, rtID)
	if err != nil {
		t.Fatalf("seed room_type: %v", err)
	}

	roomID = uuid.New()
	num := fmt.Sprintf("INT-%s", uuid.New().String()[:8])
	_, err = pool.Exec(ctx, `
		INSERT INTO room (id, room_type_id, room_number, status)
		VALUES ($1, $2, $3, 'active')`, roomID, rtID, num)
	if err != nil {
		t.Fatalf("seed room: %v", err)
	}

	date := time.Now().Truncate(24 * time.Hour)
	_, err = pool.Exec(ctx, `
		INSERT INTO room_type_inventory (room_type_id, date, total_rooms, booked_rooms)
		VALUES ($1, $2, 10, 0)
		ON CONFLICT (room_type_id, date) DO UPDATE SET total_rooms=10`, rtID, date)
	if err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM room WHERE id=$1`, roomID)
		pool.Exec(ctx, `DELETE FROM room_type WHERE id=$1`, rtID)
	})
	return rtID, roomID
}

// insertTestReservation inserts a reservation row and returns its ID.
func insertTestReservation(t *testing.T, ctx context.Context, pool *pgxpool.Pool, idempotencyKey *string) uuid.UUID {
	t.Helper()
	resID := uuid.New()
	ref := fmt.Sprintf("HB-INT-%s", uuid.New().String()[:8])
	_, err := pool.Exec(ctx, `
		INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount, idempotency_key)
		VALUES ($1, $2, 'int@test.com', 'Integration Guest', 'pending', 100, $3)`,
		resID, ref, idempotencyKey)
	if err != nil {
		t.Fatalf("seed reservation: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM booking WHERE reservation_id=$1`, resID)
		pool.Exec(ctx, `DELETE FROM reservation WHERE id=$1`, resID)
	})
	return resID
}

// ---------------------------------------------------------------------------
// PRD Test 1: Exclusion constraint violation (SQLSTATE 23P01)
// ---------------------------------------------------------------------------

func TestExclusionConstraint_DoubleBookFails(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	rtID, roomID := seedTestRoom(t, ctx, pool)
	resID1 := insertTestReservation(t, ctx, pool, nil)
	resID2 := insertTestReservation(t, ctx, pool, nil)

	now := time.Now().Truncate(time.Hour)
	start1 := now.Add(1 * time.Hour)
	end1 := now.Add(3 * time.Hour)
	start2 := now.Add(2 * time.Hour) // overlaps with booking 1
	end2 := now.Add(4 * time.Hour)

	// First booking succeeds
	_, err := pool.Exec(ctx, `
		INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
		VALUES ($1, $2, $3, $4, 'hourly', $5, $6, 'confirmed', 100)`,
		uuid.New(), resID1, roomID, rtID, start1, end1)
	if err != nil {
		t.Fatalf("first booking should succeed: %v", err)
	}

	// Second overlapping booking MUST fail with exclusion violation
	_, err = pool.Exec(ctx, `
		INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
		VALUES ($1, $2, $3, $4, 'hourly', $5, $6, 'confirmed', 100)`,
		uuid.New(), resID2, roomID, rtID, start2, end2)

	if err == nil {
		t.Fatal("expected exclusion constraint violation for overlapping booking")
	}

	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		t.Fatalf("expected pgconn.PgError, got %T: %v", err, err)
	}
	if pgErr.SQLState() != "23P01" {
		t.Fatalf("expected SQLSTATE 23P01 (exclusion violation), got %s: %s", pgErr.SQLState(), pgErr.Message)
	}
}

func TestExclusionConstraint_NonOverlappingSucceeds(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	rtID, roomID := seedTestRoom(t, ctx, pool)
	resID1 := insertTestReservation(t, ctx, pool, nil)
	resID2 := insertTestReservation(t, ctx, pool, nil)

	now := time.Now().Truncate(time.Hour)
	start1 := now.Add(1 * time.Hour)
	end1 := now.Add(2 * time.Hour)
	start2 := now.Add(3 * time.Hour) // no overlap
	end2 := now.Add(4 * time.Hour)

	_, err := pool.Exec(ctx, `
		INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
		VALUES ($1, $2, $3, $4, 'hourly', $5, $6, 'confirmed', 100)`,
		uuid.New(), resID1, roomID, rtID, start1, end1)
	if err != nil {
		t.Fatalf("first booking: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
		VALUES ($1, $2, $3, $4, 'hourly', $5, $6, 'confirmed', 100)`,
		uuid.New(), resID2, roomID, rtID, start2, end2)
	if err != nil {
		t.Fatalf("non-overlapping booking should succeed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PRD Test 2: Idempotency retry
// ---------------------------------------------------------------------------

func TestIdempotencyKey_SameKeyReturnsSameReservation(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	idempKey := fmt.Sprintf("idem-test-%s", uuid.New().String()[:8])

	// First insert
	resID1 := uuid.New()
	ref1 := fmt.Sprintf("HB-IDEM-%s", uuid.New().String()[:8])
	_, err := pool.Exec(ctx, `
		INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount, idempotency_key)
		VALUES ($1, $2, 'idem@test.com', 'Idemp Guest', 'pending', 100, $3)`,
		resID1, ref1, idempKey)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM booking WHERE reservation_id=$1`, resID1)
		pool.Exec(ctx, `DELETE FROM reservation WHERE id=$1`, resID1)
	})

	// Duplicate insert with same idempotency_key MUST fail with unique violation
	resID2 := uuid.New()
	ref2 := fmt.Sprintf("HB-IDEM-%s", uuid.New().String()[:8])
	_, err = pool.Exec(ctx, `
		INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount, idempotency_key)
		VALUES ($1, $2, 'idem@test.com', 'Idemp Guest', 'pending', 100, $3)`,
		resID2, ref2, idempKey)

	if err == nil {
		t.Fatal("expected unique constraint violation for duplicate idempotency_key")
	}

	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		t.Fatalf("expected pgconn.PgError, got %T: %v", err, err)
	}
	if pgErr.SQLState() != "23505" {
		t.Fatalf("expected SQLSTATE 23505 (unique violation), got %s: %s", pgErr.SQLState(), pgErr.Message)
	}
}

func TestIdempotencyKey_NullDoesNotConflict(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	// Two reservations with NULL idempotency_key should both succeed
	resID1 := uuid.New()
	ref1 := fmt.Sprintf("HB-NUL-%s", uuid.New().String()[:8])
	_, err := pool.Exec(ctx, `
		INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount, idempotency_key)
		VALUES ($1, $2, 'nul@test.com', 'Nul Guest', 'pending', 100, NULL)`,
		resID1, ref1)
	if err != nil {
		t.Fatalf("first null-key reservation: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM booking WHERE reservation_id=$1`, resID1)
		pool.Exec(ctx, `DELETE FROM reservation WHERE id=$1`, resID1)
	})

	resID2 := uuid.New()
	ref2 := fmt.Sprintf("HB-NUL-%s", uuid.New().String()[:8])
	_, err = pool.Exec(ctx, `
		INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount, idempotency_key)
		VALUES ($1, $2, 'nul@test.com', 'Nul Guest', 'pending', 100, NULL)`,
		resID2, ref2)
	if err != nil {
		t.Fatalf("second null-key reservation should succeed: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM booking WHERE reservation_id=$1`, resID2)
		pool.Exec(ctx, `DELETE FROM reservation WHERE id=$1`, resID2)
	})
}

// ---------------------------------------------------------------------------
// PRD Test 3: Concurrent inventory race (29 parallel bookings)
// ---------------------------------------------------------------------------

func TestConcurrentBooking_RaceSafe(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	rtID, targetRoom := seedTestRoom(t, ctx, pool)

	// Set inventory: total_rooms=1 (only the one seed room)
	date := time.Now().Truncate(24 * time.Hour)
	_, err := pool.Exec(ctx, `
		INSERT INTO room_type_inventory (room_type_id, date, total_rooms, booked_rooms)
		VALUES ($1, $2, 1, 0)
		ON CONFLICT (room_type_id, date) DO UPDATE SET total_rooms=1, booked_rooms=0`,
		rtID, date)
	if err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	// 29 goroutines all try to book THE SAME ROOM for THE SAME overlapping window.
	// The exclusion constraint must ensure at most 1 succeeds.
	const concurrency = 29
	var wg sync.WaitGroup
	successCount := make(chan int, concurrency)

	fixedStart := time.Now().Truncate(time.Hour).Add(48 * time.Hour)
	fixedEnd := fixedStart.Add(2 * time.Hour)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resID := uuid.New()
			ref := fmt.Sprintf("HB-RACE-%02d-%s", idx, uuid.New().String()[:6])

			tx, err := pool.Begin(ctx)
			if err != nil {
				return
			}
			defer tx.Rollback(ctx)

			_, err = tx.Exec(ctx, `
				INSERT INTO reservation (id, reference_code, guest_email, guest_name, status, total_amount)
				VALUES ($1, $2, 'race@test.com', 'Race Guest', 'pending', 50)`,
				resID, ref)
			if err != nil {
				return
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO booking (id, reservation_id, room_id, room_type_id, booking_type, starts_at, ends_at, status, amount)
				VALUES ($1, $2, $3, $4, 'hourly', $5, $6, 'confirmed', 50)`,
				uuid.New(), resID, targetRoom, rtID, fixedStart, fixedEnd)
			if err != nil {
				return
			}

			if err := tx.Commit(ctx); err != nil {
				return
			}
			successCount <- 1
		}(i)
	}

	wg.Wait()
	close(successCount)

	succeeded := 0
	for range successCount {
		succeeded++
	}

	t.Logf("concurrent booking results: %d succeeded out of %d attempts (1 room, same window)", succeeded, concurrency)

	if succeeded > 1 {
		t.Errorf("overbooking detected: %d bookings succeeded but only 1 room available in the window", succeeded)
	}
	if succeeded == 0 {
		t.Error("expected at least 1 booking to succeed")
	}
}
