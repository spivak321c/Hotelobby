-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- ── customer ────────────────────────────────────────────────
CREATE TABLE customer (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT   NOT NULL,
    full_name     TEXT   NOT NULL,
    phone         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── admin ───────────────────────────────────────────────────
CREATE TABLE admin (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT   NOT NULL,
    full_name     TEXT   NOT NULL,
    role          TEXT   NOT NULL CHECK (role IN ('super_admin','manager','front_desk')),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── room_type ───────────────────────────────────────────────
CREATE TABLE room_type (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    description      TEXT,
    base_rate_daily  NUMERIC(10,2) NOT NULL CHECK (base_rate_daily > 0),
    base_rate_hourly NUMERIC(10,2) CHECK (base_rate_hourly > 0),
    max_occupancy    INT NOT NULL DEFAULT 2 CHECK (max_occupancy > 0),
    is_featured      BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── room ────────────────────────────────────────────────────
CREATE TABLE room (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_type_id UUID NOT NULL REFERENCES room_type(id) ON DELETE RESTRICT,
    room_number  TEXT NOT NULL UNIQUE,
    status       TEXT NOT NULL CHECK (status IN ('active','maintenance','inactive')) DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_room_type_id ON room(room_type_id);

-- ── room_image ──────────────────────────────────────────────
CREATE TABLE room_image (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id    UUID NOT NULL REFERENCES room(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    sort_order INT  NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT false
);
CREATE UNIQUE INDEX idx_room_image_one_primary ON room_image(room_id) WHERE is_primary;

-- ── room_pricing ────────────────────────────────────────────
CREATE TABLE room_pricing (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_type_id     UUID NOT NULL REFERENCES room_type(id) ON DELETE CASCADE,
    rate_type        TEXT NOT NULL CHECK (rate_type IN ('hourly','daily')),
    rate             NUMERIC(10,2) NOT NULL CHECK (rate > 0),
    effective_range  DATERANGE NOT NULL,
    EXCLUDE USING gist (
        room_type_id WITH =,
        rate_type    WITH =,
        effective_range WITH &&
    )
);

-- ── room_type_inventory ─────────────────────────────────────
CREATE TABLE room_type_inventory (
    room_type_id UUID NOT NULL REFERENCES room_type(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    total_rooms  INT  NOT NULL CHECK (total_rooms >= 0),
    booked_rooms INT  NOT NULL DEFAULT 0 CHECK (booked_rooms >= 0 AND booked_rooms <= total_rooms),
    PRIMARY KEY (room_type_id, date)
);

-- ── reservation ─────────────────────────────────────────────
CREATE TABLE reservation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_code      TEXT NOT NULL UNIQUE,
    customer_id         UUID NULL REFERENCES customer(id) ON DELETE SET NULL,
    guest_email         CITEXT NOT NULL,
    guest_name          TEXT NOT NULL,
    guest_phone         TEXT,
    status              TEXT NOT NULL CHECK (status IN
                          ('pending','confirmed','cancelled','refunded','failed','completed')),
    total_amount        NUMERIC(10,2) NOT NULL CHECK (total_amount >= 0),
    currency            TEXT NOT NULL DEFAULT 'USD',
    created_by_admin_id UUID NULL REFERENCES admin(id) ON DELETE SET NULL,
    idempotency_key     TEXT NULL UNIQUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reservation_customer ON reservation(customer_id);
CREATE INDEX idx_reservation_lookup   ON reservation(guest_email, reference_code);
CREATE INDEX idx_reservation_admin_filter ON reservation(status, created_at);

-- ── booking ─────────────────────────────────────────────────
CREATE TABLE booking (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL REFERENCES reservation(id) ON DELETE RESTRICT,
    room_id        UUID NOT NULL REFERENCES room(id) ON DELETE RESTRICT,
    room_type_id   UUID NOT NULL REFERENCES room_type(id),
    booking_type   TEXT NOT NULL CHECK (booking_type IN ('hourly','daily')),
    starts_at      TIMESTAMPTZ NOT NULL,
    ends_at        TIMESTAMPTZ NOT NULL CHECK (ends_at > starts_at),
    status         TEXT NOT NULL CHECK (status IN
                     ('pending','confirmed','checked_in','checked_out','cancelled','refunded','failed')),
    amount         NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
    EXCLUDE USING gist (
        room_id WITH =,
        tstzrange(starts_at, ends_at) WITH &&
    ) WHERE (status IN ('pending','confirmed','checked_in'))
);
CREATE INDEX idx_booking_reservation ON booking(reservation_id);
CREATE INDEX idx_booking_room_type_date ON booking(room_type_id, starts_at);

-- ── payment ─────────────────────────────────────────────────
CREATE TABLE payment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id      UUID NOT NULL UNIQUE REFERENCES reservation(id) ON DELETE RESTRICT,
    provider            TEXT NOT NULL CHECK (provider IN ('paystack','crossmint')),
    provider_reference  TEXT NOT NULL,
    status              TEXT NOT NULL CHECK (status IN ('pending','processing','succeeded','failed','refunded')),
    amount              NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
    currency            TEXT NOT NULL,
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_reference)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS payment CASCADE;
DROP TABLE IF EXISTS booking CASCADE;
DROP TABLE IF EXISTS reservation CASCADE;
DROP TABLE IF EXISTS admin CASCADE;
DROP TABLE IF EXISTS customer CASCADE;
DROP TABLE IF EXISTS room_type_inventory CASCADE;
DROP TABLE IF EXISTS room_pricing CASCADE;
DROP TABLE IF EXISTS room_image CASCADE;
DROP TABLE IF EXISTS room CASCADE;
DROP TABLE IF EXISTS room_type CASCADE;

-- +goose StatementEnd