-- +goose Up
CREATE TABLE signing_keys (
    id VARCHAR(64) PRIMARY KEY,
    status VARCHAR(100) NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    algorithm VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ
);

CREATE INDEX signing_keys_active_fetch ON signing_keys (status, expires_at);

-- +goose Down
DROP TABLE IF EXISTS signing_keys;
