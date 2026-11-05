-- +goose Up
ALTER TABLE roles ADD COLUMN weight INTEGER;

UPDATE roles SET weight = 2147483647 WHERE internal AND name = 'superadmin';
UPDATE roles SET weight = -2147483648 WHERE internal AND name = 'default';

WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at DESC, id) AS position
    FROM roles
    WHERE weight IS NULL
)
UPDATE roles SET weight = ranked.position
FROM ranked
WHERE roles.id = ranked.id;

ALTER TABLE roles ALTER COLUMN weight SET NOT NULL;
ALTER TABLE roles ADD CONSTRAINT roles_weight_unique UNIQUE (weight) DEFERRABLE;

CREATE TABLE impersonations (
    id VARCHAR(64) PRIMARY KEY,
    impersonator_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    family_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX impersonations_unique ON impersonations (impersonator_id, application_id);

ALTER TABLE sessions ADD COLUMN impersonator_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE authorization_codes ADD COLUMN impersonator_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX sessions_impersonator_lookup ON sessions (impersonator_id) WHERE impersonator_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS sessions_impersonator_lookup;
ALTER TABLE authorization_codes DROP COLUMN impersonator_id;
ALTER TABLE sessions DROP COLUMN impersonator_id;
DROP TABLE IF EXISTS impersonations;
ALTER TABLE roles DROP CONSTRAINT roles_weight_unique;
ALTER TABLE roles DROP COLUMN weight;
