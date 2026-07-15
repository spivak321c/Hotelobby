-- +goose Up
-- +goose StatementBegin

ALTER TABLE reservation
    ADD COLUMN IF NOT EXISTS cancellation_reason TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE reservation
    DROP COLUMN IF EXISTS cancellation_reason;

-- +goose StatementEnd
