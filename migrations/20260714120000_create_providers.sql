-- +goose Up
CREATE TABLE providers (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(64) NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(512) NOT NULL,
    scopes VARCHAR(128)[] NOT NULL,
    oidc_url VARCHAR(255),
    authorization_url VARCHAR(255),
    token_url VARCHAR(255),
    userinfo_url VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS providers;
