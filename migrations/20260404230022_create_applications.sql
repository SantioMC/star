-- +goose Up
CREATE TABLE applications (
    id VARCHAR(64) PRIMARY KEY,
    kind VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    application_url VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255),
    internal BOOLEAN NOT NULL DEFAULT FALSE,
    token_duration INTEGER NOT NULL DEFAULT 3600,
    session_duration INTEGER NOT NULL DEFAULT 604800,
    redirect_uris VARCHAR(255)[] NOT NULL,
    development_mode BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
);

ALTER TABLE sessions ADD COLUMN application_id VARCHAR(64) REFERENCES applications(id) ON DELETE CASCADE;

CREATE TABLE authorization_codes (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    code VARCHAR(255) NOT NULL,
    redirect_uri VARCHAR(255) NOT NULL,
    scopes VARCHAR(128)[] NOT NULL,
    state VARCHAR(255),
    nonce VARCHAR(255),
    code_challenge VARCHAR(255),
    code_challenge_method VARCHAR(12),
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    redeemed_at TIMESTAMP
);

-- +goose Down
ALTER TABLE sessions DROP COLUMN application_id;
DROP TABLE IF EXISTS applications;