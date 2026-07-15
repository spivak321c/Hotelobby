-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin

-- Replace room(room_type_id) with composite (room_type_id, status)
-- for filtered queries by both columns
DROP INDEX IF EXISTS idx_room_type_id;
CREATE INDEX IF NOT EXISTS idx_room_type_status
ON room(room_type_id, status);

-- Covering index for CountActiveBookings: room_id + status + ends_at range scan
CREATE INDEX IF NOT EXISTS idx_booking_room_active_ends
ON booking(room_id, status, ends_at);

-- Covering index for nightly inventory rebuild subquery:
-- room_type_id + status + starts_at + ends_at date-range filter
CREATE INDEX IF NOT EXISTS idx_booking_inventory_rebuild
ON booking(room_type_id, status, starts_at, ends_at);

-- B-tree index for sorted pricing lookups by lower(effective_range)
-- GiST exclusion index does not support ORDER BY lower()
CREATE INDEX IF NOT EXISTS idx_room_pricing_type_lower
ON room_pricing(room_type_id, lower(effective_range));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_room_pricing_type_lower;
DROP INDEX IF EXISTS idx_booking_inventory_rebuild;
DROP INDEX IF EXISTS idx_booking_room_active_ends;
DROP INDEX IF EXISTS idx_room_type_status;
CREATE INDEX IF NOT EXISTS idx_room_type_id ON room(room_type_id);

-- +goose StatementEnd
