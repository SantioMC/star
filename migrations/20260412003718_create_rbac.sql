-- +goose Up
CREATE TABLE roles (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    internal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE role_bindings (
    id VARCHAR(64) PRIMARY KEY,
    role_id VARCHAR(64) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE permissions (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE permission_bindings (
    id VARCHAR(64) PRIMARY KEY,
    permission_id VARCHAR(64) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    role_id VARCHAR(64) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX roles_name_unique ON roles (name);
CREATE UNIQUE INDEX role_bindings_unique ON role_bindings (role_id, user_id);
CREATE INDEX role_bindings_user_lookup ON role_bindings (user_id);
CREATE UNIQUE INDEX permissions_name_unique ON permissions (name);
CREATE UNIQUE INDEX permission_bindings_unique ON permission_bindings (role_id, permission_id);

-- +goose Down
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS permission_bindings;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS role_bindings;
