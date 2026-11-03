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
    trust_email BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE accounts ADD COLUMN provider_id VARCHAR(64) REFERENCES providers(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX accounts_provider_lookup ON accounts (provider_id, external_id);

ALTER TABLE login_flows ADD COLUMN account_id VARCHAR(64) REFERENCES accounts(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE login_flows DROP COLUMN account_id;
DROP INDEX IF EXISTS accounts_provider_lookup;
ALTER TABLE accounts DROP COLUMN provider_id;
DROP TABLE IF EXISTS providers;
