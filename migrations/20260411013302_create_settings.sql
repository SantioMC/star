-- +goose Up
CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    allow_registration BOOLEAN NOT NULL DEFAULT TRUE,
    allow_logins BOOLEAN NOT NULL DEFAULT TRUE,
    email_code_length INTEGER NOT NULL DEFAULT 8,
    audit_retention_days INTEGER,
    audit_anonymize BOOLEAN NOT NULL DEFAULT FALSE,
    password_settings JSONB NOT NULL
);

CREATE TABLE branding (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    accent_color VARCHAR(7) NOT NULL,
    heading_font VARCHAR(64) NOT NULL,
    body_font VARCHAR(64) NOT NULL,
    layout VARCHAR(16) NOT NULL,
    support_email VARCHAR(255),
    privacy_url VARCHAR(2048),
    terms_url VARCHAR(2048),
    logo BYTEA,
    logo_type VARCHAR(32),
    logo_updated_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS branding;
DROP TABLE IF EXISTS settings;
