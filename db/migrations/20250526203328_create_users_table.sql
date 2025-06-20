-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(254) NOT NULL UNIQUE,
    phone VARCHAR(30) NOT NULL UNIQUE,
    password VARCHAR(254) NOT NULL,
    preferences JSONB,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users (email)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_users_phone ON users (phone)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE INDEX idx_users_archived_at ON users (archived_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_archived_at;

DROP INDEX IF EXISTS idx_users_deleted_at;

DROP INDEX IF EXISTS idx_users_phone;

DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS users;

-- +goose StatementEnd
