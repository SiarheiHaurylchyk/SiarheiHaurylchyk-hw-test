CREATE TABLE IF NOT EXISTS events (
    id            TEXT PRIMARY KEY,
    title         TEXT        NOT NULL,
    start_time    TIMESTAMPTZ NOT NULL,
    end_time      TIMESTAMPTZ NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    user_id       TEXT        NOT NULL,
    notify_before BIGINT      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS events_user_id_start_time_idx ON events (user_id, start_time);
