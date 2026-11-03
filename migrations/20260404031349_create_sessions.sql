-- +goose Up
CREATE TABLE sessions (
    id VARCHAR(64) PRIMARY KEY,
    family_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(64) REFERENCES accounts(id) ON DELETE CASCADE,
    ip_address VARCHAR(100) NOT NULL,
    user_agent VARCHAR(255) NOT NULL,
    geolocation VARCHAR(255),
    token VARCHAR(255) NOT NULL,
    scopes VARCHAR(128)[],
    device VARCHAR(64) NOT NULL,
    status VARCHAR(100) NOT NULL,
    authenticated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_token_lookup ON sessions (token);
CREATE INDEX sessions_family_lookup ON sessions (family_id);
CREATE INDEX sessions_user_lookup ON sessions (user_id);
CREATE INDEX sessions_expiry_lookup ON sessions (expires_at);

CREATE TABLE token_revocations (
    subject VARCHAR(64) PRIMARY KEY,
    kind VARCHAR(32) NOT NULL,
    user_id VARCHAR(64),
    reason VARCHAR(64) NOT NULL,
    revoked_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX token_revocations_revoked_lookup ON token_revocations (revoked_at);
CREATE INDEX token_revocations_expiry_lookup ON token_revocations (expires_at);

-- +goose Down
DROP TABLE IF EXISTS token_revocations;
DROP TABLE IF EXISTS sessions;
