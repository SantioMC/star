-- +goose Up
CREATE INDEX user_emails_email_search ON user_emails USING GIN (email gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS user_emails_email_search;
