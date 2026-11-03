-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),    
    profile_picture VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_emails (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(64) NOT NULL,
    code_attempts INTEGER NOT NULL DEFAULT 0,
    code_expires_at TIMESTAMPTZ,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    preferred BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE accounts (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    type VARCHAR(100) NOT NULL,

    password VARCHAR(255),
    password_alg VARCHAR(100),

    external_id VARCHAR(255),
    account_name VARCHAR(255),
    account_email VARCHAR(255),
    account_avatar VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX users_name_unique ON users (LOWER(name));
CREATE INDEX users_name_search ON users USING GIN (name gin_trgm_ops);
CREATE INDEX user_emails_user_lookup ON user_emails (user_id);
CREATE INDEX user_emails_email_lookup ON user_emails (LOWER(email));
CREATE UNIQUE INDEX user_emails_verified_unique ON user_emails (LOWER(email)) WHERE verified;
CREATE INDEX accounts_user_lookup ON accounts (user_id);

CREATE TABLE password_resets (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX password_resets_token_lookup ON password_resets (token);
CREATE UNIQUE INDEX password_resets_user_lookup ON password_resets (user_id);

-- +goose Down
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS user_emails;
DROP TABLE IF EXISTS users;