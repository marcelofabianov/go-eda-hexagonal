-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_version VARCHAR(50) NOT NULL,
    event_context JSONB NOT NULL,
    trace_id UUID NOT NULL,
    user_author_id UUID,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_event_type ON audit_logs (event_type);

CREATE INDEX idx_audit_logs_trace_id ON audit_logs (trace_id);

CREATE INDEX idx_audit_logs_user_author_id ON audit_logs (user_author_id);

CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_audit_logs_created_at;

DROP INDEX IF EXISTS idx_audit_logs_user_author_id;

DROP INDEX IF EXISTS idx_audit_logs_trace_id;

DROP INDEX IF EXISTS idx_audit_logs_event_type;

DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
