-- +goose Up
CREATE TABLE user_invites (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL,
    token VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX user_invites_token_lookup ON user_invites (token);
CREATE UNIQUE INDEX user_invites_user_lookup ON user_invites (user_id);
CREATE INDEX user_invites_expiry_lookup ON user_invites (expires_at);

-- +goose Down
DROP TABLE IF EXISTS user_invites;
