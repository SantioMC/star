-- +goose Up
CREATE TABLE applications (
    id VARCHAR(64) PRIMARY KEY,
    kind VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(32) NOT NULL UNIQUE,
    application_url VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret VARCHAR(255),
    internal BOOLEAN NOT NULL DEFAULT FALSE,
    consent VARCHAR(16) NOT NULL DEFAULT 'default',
    token_duration INTEGER NOT NULL DEFAULT 3600,
    session_duration INTEGER NOT NULL DEFAULT 604800,
    redirect_uris VARCHAR(255)[] NOT NULL,
    development_mode BOOLEAN NOT NULL DEFAULT FALSE,
    privileged_scopes BOOLEAN NOT NULL DEFAULT FALSE,
    signing_algorithm VARCHAR(16) NOT NULL DEFAULT 'ES512',
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT applications_token_duration_bounds CHECK (token_duration BETWEEN 60 AND 3600),
    CONSTRAINT applications_session_duration_bounds CHECK (session_duration > token_duration)
);

CREATE TABLE application_scopes (
    id VARCHAR(64) PRIMARY KEY,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL,
    requestable BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX application_scopes_application_lookup ON application_scopes (application_id);

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
    family_id VARCHAR(64),
    authenticated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    redeemed_at TIMESTAMPTZ
);

CREATE INDEX authorization_codes_code_lookup ON authorization_codes (code);
CREATE INDEX authorization_codes_user_lookup ON authorization_codes (user_id, application_id);

CREATE TABLE application_grants (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id VARCHAR(64) NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    scopes VARCHAR(128)[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, application_id)
);

-- +goose Down
DROP TABLE IF EXISTS application_grants;
DROP TABLE IF EXISTS authorization_codes;
ALTER TABLE sessions DROP COLUMN application_id;
DROP TABLE IF EXISTS application_scopes;
DROP TABLE IF EXISTS applications;
