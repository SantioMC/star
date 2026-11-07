-- +goose Up
ALTER TABLE providers ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX accounts_user_provider ON accounts (user_id, provider_id) WHERE provider_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS accounts_user_provider;

ALTER TABLE providers DROP COLUMN enabled;
