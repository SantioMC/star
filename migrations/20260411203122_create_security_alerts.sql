-- +goose Up
CREATE TABLE security_alerts (
    id VARCHAR(64) PRIMARY KEY,
    kind VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(32) NOT NULL,
    data JSONB,
    user_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE,
    application_id VARCHAR(64) REFERENCES applications(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS security_alerts;
