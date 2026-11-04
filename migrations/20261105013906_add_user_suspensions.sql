-- +goose Up
CREATE TABLE user_suspensions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(500) NOT NULL,
    suspended_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    lifted_at TIMESTAMPTZ,
    lifted_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL,
    lift_reason VARCHAR(500)
);

CREATE INDEX user_suspensions_user_lookup ON user_suspensions (user_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS user_suspensions;
