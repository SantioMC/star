-- +goose Up
ALTER TABLE applications ADD COLUMN device_flow BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE device_authorizations (
    id VARCHAR(64) PRIMARY KEY,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    user_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE,
    device_code VARCHAR(64) NOT NULL,
    user_code VARCHAR(64) NOT NULL,
    scopes VARCHAR(128)[] NOT NULL,
    status VARCHAR(16) NOT NULL,
    device VARCHAR(64) NOT NULL,
    geolocation VARCHAR(255),
    poll_interval INTEGER NOT NULL,
    family_id VARCHAR(64),
    authenticated_at TIMESTAMPTZ,
    last_polled_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    redeemed_at TIMESTAMPTZ,
    CONSTRAINT device_authorizations_approved_identity CHECK (status <> 'approved' OR (user_id IS NOT NULL AND authenticated_at IS NOT NULL))
);

CREATE UNIQUE INDEX device_authorizations_device_code_lookup ON device_authorizations (device_code);
CREATE UNIQUE INDEX device_authorizations_user_code_lookup ON device_authorizations (user_code);
CREATE INDEX device_authorizations_user_lookup ON device_authorizations (user_id, application_id);
CREATE INDEX device_authorizations_expiry_lookup ON device_authorizations (expires_at);

-- +goose Down
DROP TABLE IF EXISTS device_authorizations;
ALTER TABLE applications DROP COLUMN device_flow;
