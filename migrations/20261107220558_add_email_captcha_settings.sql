-- +goose Up
ALTER TABLE settings
    ADD COLUMN email_provider VARCHAR(16) NOT NULL DEFAULT 'disabled',
    ADD COLUMN email_sender VARCHAR(320) NOT NULL DEFAULT '',
    ADD COLUMN smtp_host VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN smtp_port INTEGER NOT NULL DEFAULT 587,
    ADD COLUMN smtp_username VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN smtp_password TEXT NOT NULL DEFAULT '',
    ADD COLUMN resend_api_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN captcha_provider VARCHAR(16) NOT NULL DEFAULT 'disabled',
    ADD COLUMN turnstile_site_key VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN turnstile_secret_key TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE settings
    DROP COLUMN turnstile_secret_key,
    DROP COLUMN turnstile_site_key,
    DROP COLUMN captcha_provider,
    DROP COLUMN resend_api_key,
    DROP COLUMN smtp_password,
    DROP COLUMN smtp_username,
    DROP COLUMN smtp_port,
    DROP COLUMN smtp_host,
    DROP COLUMN email_sender,
    DROP COLUMN email_provider;
