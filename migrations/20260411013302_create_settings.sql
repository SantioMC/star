-- +goose Up
CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    allow_registration BOOLEAN NOT NULL DEFAULT TRUE,
    allow_logins BOOLEAN NOT NULL DEFAULT TRUE,
    email_code_length INTEGER NOT NULL DEFAULT 8,
    password_settings JSONB NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS settings;
