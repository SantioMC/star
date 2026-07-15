-- +goose Up
CREATE TABLE sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(64) REFERENCES accounts(id) ON DELETE CASCADE,
    ip_address VARCHAR(100) NOT NULL,
    user_agent VARCHAR(255) NOT NULL,
    geolocation VARCHAR(255),
    token VARCHAR(255) NOT NULL,
    device VARCHAR(64) NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
