-- +goose Up
CREATE TABLE totp (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(100) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE backup_codes (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(32) NOT NULL
);

-- +goose Down
DROP TABLE totp;
DROP TABLE backup_codes;
