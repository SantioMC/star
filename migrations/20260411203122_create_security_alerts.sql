-- +goose Up
CREATE TABLE security_alerts (
    id VARCHAR(64) PRIMARY KEY,
    kind VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(32) NOT NULL,
    data JSONB,
    user_id VARCHAR(64) REFERENCES users(id) ON DELETE CASCADE,
    application_id VARCHAR(64) REFERENCES applications(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    resolved_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX security_alerts_unresolved ON security_alerts (created_at DESC) WHERE resolved_at IS NULL;
CREATE INDEX security_alerts_user_lookup ON security_alerts (user_id);

-- +goose Down
DROP TABLE IF EXISTS security_alerts;
