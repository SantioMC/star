-- +goose Up
ALTER TABLE users ADD COLUMN type VARCHAR(16) NOT NULL DEFAULT 'regular';

CREATE INDEX users_type_lookup ON users (type, created_at DESC) WHERE type <> 'regular';

-- +goose Down
DROP INDEX IF EXISTS users_type_lookup;
ALTER TABLE users DROP COLUMN type;
