-- +goose Up
ALTER TABLE settings
    ADD COLUMN lockout_threshold INTEGER NOT NULL DEFAULT 10,
    ADD COLUMN lockout_window_minutes INTEGER NOT NULL DEFAULT 5,
    ADD COLUMN lockout_duration_minutes INTEGER NOT NULL DEFAULT 60;

ALTER TABLE login_attempts
    ADD COLUMN window_started_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE login_attempts SET window_started_at = last_attempt_at;

-- +goose Down
ALTER TABLE login_attempts
    DROP COLUMN window_started_at;

ALTER TABLE settings
    DROP COLUMN lockout_duration_minutes,
    DROP COLUMN lockout_window_minutes,
    DROP COLUMN lockout_threshold;
