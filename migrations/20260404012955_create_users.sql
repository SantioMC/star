-- +goose Up
CREATE TABLE users (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),    
    profile_picture VARCHAR(255),
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE user_emails (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(16) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    preferred BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
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

    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS user_emails;
DROP TABLE IF EXISTS users;