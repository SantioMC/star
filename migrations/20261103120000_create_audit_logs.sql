-- +goose Up
CREATE TABLE audit_logs (
    id VARCHAR(64) NOT NULL,
    action VARCHAR(64) NOT NULL,
    actor_id VARCHAR(64),
    actor_name TEXT,
    target_type VARCHAR(32),
    target_id VARCHAR(64),
    target_name TEXT,
    application_id VARCHAR(64),
    provider_id VARCHAR(64),
    ip_address TEXT,
    ip_hash TEXT,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;

CREATE INDEX audit_logs_actor_lookup ON audit_logs (actor_id, created_at DESC, id DESC) WHERE actor_id IS NOT NULL;
CREATE INDEX audit_logs_target_lookup ON audit_logs (target_id, created_at DESC, id DESC) WHERE target_id IS NOT NULL;
CREATE INDEX audit_logs_action_lookup ON audit_logs (action, created_at DESC, id DESC);
CREATE INDEX audit_logs_application_lookup ON audit_logs (application_id, created_at DESC, id DESC) WHERE application_id IS NOT NULL;
CREATE INDEX audit_logs_provider_lookup ON audit_logs (provider_id, created_at DESC, id DESC) WHERE provider_id IS NOT NULL;
CREATE INDEX audit_logs_ip_lookup ON audit_logs (ip_hash, created_at DESC, id DESC) WHERE ip_hash IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
