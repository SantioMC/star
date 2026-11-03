-- +goose Up
CREATE TABLE totp (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_used_step BIGINT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE backup_codes (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX backup_codes_lookup ON backup_codes (user_id, code);
CREATE INDEX totp_user_lookup ON totp (user_id);

CREATE TABLE login_flows (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL,
    stage VARCHAR(32) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX login_flows_token_lookup ON login_flows (token);
CREATE INDEX login_flows_user_lookup ON login_flows (user_id);

CREATE TABLE login_attempts (
    user_id VARCHAR(64) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    attempts INTEGER NOT NULL,
    last_attempt_at TIMESTAMPTZ NOT NULL,
    locked_until TIMESTAMPTZ
);

CREATE TABLE passkeys (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    credential_id BYTEA NOT NULL UNIQUE,
    credential JSONB NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX passkeys_user ON passkeys (user_id);

CREATE TABLE passkey_challenges (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE,
    login_flow_id VARCHAR(64) REFERENCES login_flows(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    kind VARCHAR(32) NOT NULL,
    session JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE passkey_challenges;
DROP TABLE passkeys;
DROP TABLE login_attempts;
DROP TABLE login_flows;
DROP TABLE totp;
DROP TABLE backup_codes;
