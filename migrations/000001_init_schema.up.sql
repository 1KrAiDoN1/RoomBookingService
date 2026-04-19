CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL CHECK (role IN ('admin', 'user')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id          UUID        PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    capacity    INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS schedules (
    id           UUID  PRIMARY KEY,
    room_id      UUID  NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week INT[] NOT NULL,
    start_time   TEXT  NOT NULL,
    end_time     TEXT  NOT NULL
);


CREATE TABLE IF NOT EXISTS slots (
    id      UUID        PRIMARY KEY,
    room_id UUID        NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at   TIMESTAMPTZ NOT NULL,
    end_at   TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_slots_room_start UNIQUE (room_id, start_at)
);


CREATE INDEX IF NOT EXISTS idx_slots_room_start ON slots (room_id, start_at);


CREATE TABLE IF NOT EXISTS bookings (
    id              UUID        PRIMARY KEY,
    slot_id         UUID        NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status          TEXT        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    conference_link TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bookings_active_slot ON bookings (slot_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings (user_id);

INSERT INTO users (id, email, password_hash, role) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@example.com', 'dummy_hash_admin', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'user1@example.com', 'dummy_hash_user', 'user')
ON CONFLICT (id) DO NOTHING;


CREATE INDEX idx_bookings_created_at ON bookings(created_at DESC);