package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

const RoleKey = "rol"
const PermissionKey = "perm"
const RoleBindingKey = "rb"
const PermissionBindingKey = "pb"

type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Internal    bool      `json:"internal"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type RoleBinding struct {
	ID        string    `json:"id"`
	RoleID    string    `json:"role_id" db:"role_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PermissionBinding struct {
	ID           string    `json:"id"`
	PermissionID string    `json:"permission_id" db:"permission_id"`
	RoleID       string    `json:"role_id" db:"role_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

func (db *Database) GetRoleByID(ctx context.Context, id string) *Role {
	role, err := sqlx.OneContext[Role](
		ctx,
		db.DB(),
		`SELECT * FROM roles WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get role by id", "err", err)
		return nil
	}

	return &role
}

func (db *Database) GetRoleByName(ctx context.Context, name string) *Role {
	role, err := sqlx.OneContext[Role](
		ctx,
		db.DB(),
		`SELECT * FROM roles WHERE name = $1 LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get role by name", "err", err)
		return nil
	}

	return &role
}

func (db *Database) GetPermissionByID(ctx context.Context, id string) *Permission {
	permission, err := sqlx.OneContext[Permission](
		ctx,
		db.DB(),
		`SELECT * FROM permissions WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get permission by id", "err", err)
		return nil
	}

	return &permission
}

func (db *Database) GetPermissionByName(ctx context.Context, name string) *Permission {
	permission, err := sqlx.OneContext[Permission](
		ctx,
		db.DB(),
		`SELECT * FROM permissions WHERE name = $1 LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get permission by name", "err", err)
		return nil
	}

	return &permission
}

func (db *Database) CreateRole(ctx context.Context, role *Role) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO roles (
			id, name, description, internal, created_at
		) VALUES ($1, $2, $3, $4, $5)
		`,
		role.ID,
		role.Name,
		role.Description,
		role.Internal,
		role.CreatedAt,
	)

	return err
}

func (db *Database) UpdateRole(ctx context.Context, role *Role) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE roles
		SET name = $2, description = $3
		WHERE id = $1
		`,
		role.ID,
		role.Name,
		role.Description,
	)

	return err
}

func (db *Database) DeleteRoleByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM roles WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CreatePermission(ctx context.Context, permission *Permission) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO permissions (
			id, name, description, created_at
		) VALUES ($1, $2, $3, $4)
		`,
		permission.ID,
		permission.Name,
		permission.Description,
		permission.CreatedAt,
	)

	return err
}

func (db *Database) DeletePermissionByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM permissions WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CreateRoleBinding(ctx context.Context, binding *RoleBinding) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO role_bindings (
			id, role_id, user_id, created_at
		) VALUES ($1, $2, $3, $4)
		`,
		binding.ID,
		binding.RoleID,
		binding.UserID,
		binding.CreatedAt,
	)

	return err
}

func (db *Database) DeleteRoleBindingByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM role_bindings WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CreatePermissionBinding(ctx context.Context, binding *PermissionBinding) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO permission_bindings (
			id, permission_id, role_id, created_at
		) VALUES ($1, $2, $3, $4)
		`,
		binding.ID,
		binding.PermissionID,
		binding.RoleID,
		binding.CreatedAt,
	)

	return err
}

func (db *Database) DeletePermissionBindingByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM permission_bindings WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) GetUserPermissions(ctx context.Context, userId string) []Permission {
	permissions, err := sqlx.ListContext[Permission](
		ctx,
		db.DB(),
		`
		SELECT permissions.* FROM permissions
		INNER JOIN permission_bindings ON permissions.id = permission_bindings.permission_id
		INNER JOIN role_bindings ON permission_bindings.role_id = role_bindings.role_id
		WHERE role_bindings.user_id = $1
		`,
		userId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Permission{}
	} else if err != nil {
		slog.Error("failed to get user permissions", "err", err)
		return []Permission{}
	}

	return permissions
}

func (db *Database) UserHasPermission(ctx context.Context, userId string, permission string) bool {
	exists, err := sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`
		SELECT EXISTS (
			SELECT 1 FROM permissions
			INNER JOIN permission_bindings ON permissions.id = permission_bindings.permission_id
			INNER JOIN role_bindings ON permission_bindings.role_id = role_bindings.role_id
			WHERE role_bindings.user_id = $1 AND permissions.name = $2
		)
		`,
		userId,
		permission,
	)

	if err != nil {
		// Fail closed
		slog.Error("failed to check user permission", "err", err, "user", userId, "permission", permission)
		return false
	}

	return exists
}

func (db *Database) GetUserRoles(ctx context.Context, userId string) []Role {
	roles, err := sqlx.ListContext[Role](
		ctx,
		db.DB(),
		`
		SELECT roles.* FROM roles
		INNER JOIN role_bindings ON roles.id = role_bindings.role_id
		WHERE role_bindings.user_id = $1
		`,
		userId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Role{}
	} else if err != nil {
		slog.Error("failed to get user roles", "err", err)
		return []Role{}
	}

	return roles
}

func (db *Database) GetRoles(ctx context.Context) []Role {
	roles, err := sqlx.ListContext[Role](
		ctx,
		db.DB(),
		`SELECT * FROM roles ORDER BY created_at DESC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Role{}
	} else if err != nil {
		slog.Error("failed to get roles", "err", err)
		return []Role{}
	}

	return roles
}

func (db *Database) GetPermissions(ctx context.Context) []Permission {
	permissions, err := sqlx.ListContext[Permission](
		ctx,
		db.DB(),
		`SELECT * FROM permissions ORDER BY created_at DESC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Permission{}
	} else if err != nil {
		slog.Error("failed to get permissions", "err", err)
		return []Permission{}
	}

	return permissions
}

func (db *Database) GetRolePermissions(ctx context.Context, roleID string) []Permission {
	permissions, err := sqlx.ListContext[Permission](
		ctx,
		db.DB(),
		`
		SELECT permissions.* FROM permissions
		INNER JOIN permission_bindings ON permissions.id = permission_bindings.permission_id
		WHERE permission_bindings.role_id = $1
		ORDER BY permissions.name
		`,
		roleID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Permission{}
	} else if err != nil {
		slog.Error("failed to get role permissions", "err", err)
		return []Permission{}
	}

	return permissions
}

func (db *Database) DeletePermissionBindingsByRoleID(ctx context.Context, roleID string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM permission_bindings WHERE role_id = $1`,
		roleID,
	)

	return err
}

func (db *Database) GetUserRoleBindings(ctx context.Context, userID string) []RoleBinding {
	bindings, err := sqlx.ListContext[RoleBinding](
		ctx,
		db.DB(),
		`SELECT * FROM role_bindings WHERE user_id = $1`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []RoleBinding{}
	} else if err != nil {
		slog.Error("failed to get user role bindings", "err", err)
		return []RoleBinding{}
	}

	return bindings
}

func (db *Database) CountRoleBindingsByRoleID(ctx context.Context, roleID string) (int, error) {
	count, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM role_bindings WHERE role_id = $1`,
		roleID,
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *Database) CountRoles(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM roles`)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *Database) CountPermissions(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM permissions`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
