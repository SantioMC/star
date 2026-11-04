-- +goose Up
ALTER TABLE settings
    ADD COLUMN email_verification BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN email_domain_mode VARCHAR(16) NOT NULL DEFAULT 'disabled',
    ADD COLUMN email_domains JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN block_disposable_emails BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE user_emails ADD COLUMN self_asserted BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE user_emails DROP COLUMN self_asserted;

ALTER TABLE settings
    DROP COLUMN block_disposable_emails,
    DROP COLUMN email_domains,
    DROP COLUMN email_domain_mode,
    DROP COLUMN email_verification;
