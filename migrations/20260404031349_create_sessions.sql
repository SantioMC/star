-- +goose Up
CREATE TABLE sessions (
    id VARCHAR(64) PRIMARY KEY,
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

-- +goose Down
DROP TABLE IF EXISTS sessions;
