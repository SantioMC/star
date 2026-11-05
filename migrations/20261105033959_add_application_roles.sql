-- +goose Up
ALTER TABLE applications ADD COLUMN role_restricted BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE application_role_bindings (
    id VARCHAR(64) PRIMARY KEY,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    role_id VARCHAR(64) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX application_role_bindings_unique ON application_role_bindings (application_id, role_id);
CREATE INDEX application_role_bindings_role_lookup ON application_role_bindings (role_id);

-- +goose Down
DROP TABLE IF EXISTS application_role_bindings;
ALTER TABLE applications DROP COLUMN role_restricted;
