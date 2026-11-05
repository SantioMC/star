package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/lib/pq"
	"github.com/vinovest/sqlx"
)

const RoleKey = "rol"
const PermissionKey = "perm"
const RoleBindingKey = "rb"
const PermissionBindingKey = "pb"

const SuperadminRoleName = "superadmin"
const DefaultRoleName = "default"

const SuperadminRoleWeight = math.MaxInt32
const DefaultRoleWeight = math.MinInt32

var ErrRoleOrderOutdated = errors.New("the role order does not match the current roles")
var ErrRoleOrderForbidden = errors.New("the role order moves a role outside of the requester's reach")

type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Internal    bool      `json:"internal"`
	Weight      int       `json:"weight"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func (role *Role) IsDefault() bool {
	return role.Internal && role.Name == DefaultRoleName
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
	return db.insertRole(ctx, db.DB(), role)
}

func (db *Database) insertRole(ctx context.Context, execer sqlx.ExecerContext, role *Role) error {
	_, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO roles (
			id, name, description, internal, weight, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		`,
		role.ID,
		role.Name,
		role.Description,
		role.Internal,
		role.Weight,
		role.CreatedAt,
	)

	return err
}

func (db *Database) UpdateRole(ctx context.Context, role *Role) error {
	return db.updateRole(ctx, db.DB(), role)
}

func (db *Database) updateRole(ctx context.Context, execer sqlx.ExecerContext, role *Role) error {
	_, err := execer.ExecContext(
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
	return db.insertRoleBinding(ctx, db.DB(), binding)
}

func (db *Database) insertRoleBinding(ctx context.Context, execer sqlx.ExecerContext, binding *RoleBinding) error {
	_, err := execer.ExecContext(
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
	return db.insertPermissionBinding(ctx, db.DB(), binding)
}

func (db *Database) insertPermissionBinding(ctx context.Context, execer sqlx.ExecerContext, binding *PermissionBinding) error {
	_, err := execer.ExecContext(
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
	permissions, err := list[Permission](
		ctx,
		db.DB(),
		`
		SELECT DISTINCT permissions.* FROM permissions
		INNER JOIN permission_bindings ON permissions.id = permission_bindings.permission_id
		WHERE permission_bindings.role_id IN (
			SELECT role_id FROM role_bindings WHERE user_id = $1
			UNION
			SELECT id FROM roles WHERE name = $2 AND internal
		)
		`,
		userId,
		DefaultRoleName,
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
			WHERE permissions.name = $2 AND permission_bindings.role_id IN (
				SELECT role_id FROM role_bindings WHERE user_id = $1
				UNION
				SELECT id FROM roles WHERE name = $3 AND internal
			)
		)
		`,
		userId,
		permission,
		DefaultRoleName,
	)

	if err != nil {
		// Fail closed
		slog.Error("failed to check user permission", "err", err, "user", userId, "permission", permission)
		return false
	}

	return exists
}

func (db *Database) GetUserRoles(ctx context.Context, userId string) []Role {
	roles, err := list[Role](
		ctx,
		db.DB(),
		`
		SELECT roles.* FROM roles
		INNER JOIN role_bindings ON roles.id = role_bindings.role_id
		WHERE role_bindings.user_id = $1
		ORDER BY roles.weight DESC
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

func (db *Database) GetUserWeight(ctx context.Context, userID string) (int, error) {
	return sqlx.OneContext[int](
		ctx,
		db.DB(),
		`
		SELECT COALESCE(MAX(weight), $2::INTEGER) FROM roles
		WHERE id IN (SELECT role_id FROM role_bindings WHERE user_id = $1)
		OR (name = $3 AND internal)
		`,
		userID,
		DefaultRoleWeight,
		DefaultRoleName,
	)
}

func (db *Database) GetRoles(ctx context.Context) []Role {
	roles, err := list[Role](
		ctx,
		db.DB(),
		`SELECT * FROM roles ORDER BY weight DESC`,
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
	permissions, err := list[Permission](
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
	permissions, err := list[Permission](
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
	bindings, err := list[RoleBinding](
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

func (db *Database) insertPermissionBindings(ctx context.Context, execer sqlx.ExecerContext, roleID string, permissions []Permission) error {
	for _, permission := range permissions {
		binding := &PermissionBinding{
			ID:           CreateID(PermissionBindingKey),
			PermissionID: permission.ID,
			RoleID:       roleID,
			CreatedAt:    time.Now(),
		}

		if err := db.insertPermissionBinding(ctx, execer, binding); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) CreateRoleWithPermissions(ctx context.Context, role *Role, permissions []Permission) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	weight, err := sqlx.OneContext[int](ctx, tx, `SELECT COALESCE(MIN(weight) - 1, 0) FROM roles WHERE NOT internal`)
	if err != nil {
		return err
	}

	role.Weight = weight

	if err := db.insertRole(ctx, tx, role); err != nil {
		return duplicate(err)
	}

	if err := db.insertPermissionBindings(ctx, tx, role.ID, permissions); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) UpdateRoleWithPermissions(ctx context.Context, role *Role, permissions []Permission) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := db.updateRole(ctx, tx, role); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM permission_bindings WHERE role_id = $1`, role.ID); err != nil {
		return err
	}

	if err := db.insertPermissionBindings(ctx, tx, role.ID, permissions); err != nil {
		return err
	}

	if role.IsDefault() {
		err = db.revokeAll(ctx, tx, RevocationReasonRolesChanged)
	} else {
		err = db.revokeRole(ctx, tx, role.ID, RevocationReasonRolesChanged)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) DeleteRoleWithRevocation(ctx context.Context, role *Role) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := db.revokeRole(ctx, tx, role.ID, RevocationReasonRolesChanged); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, role.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) ReorderRoles(ctx context.Context, roleIDs []string, reach int) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	current, err := list[Role](ctx, tx, `SELECT * FROM roles WHERE NOT internal ORDER BY weight DESC FOR UPDATE`)
	if err != nil {
		return err
	}

	if len(current) != len(roleIDs) {
		return ErrRoleOrderOutdated
	}

	previous := make(map[string]int, len(current))
	for _, role := range current {
		previous[role.ID] = role.Weight
	}

	weights := make([]int64, len(roleIDs))
	for position, roleID := range roleIDs {
		weight, known := previous[roleID]
		if !known {
			return ErrRoleOrderOutdated
		}

		delete(previous, roleID)

		next := current[position].Weight
		if weight != next && (weight > reach || next > reach) {
			return ErrRoleOrderForbidden
		}

		weights[position] = int64(next)
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE roles SET weight = ordered.weight
		FROM unnest($1::VARCHAR[], $2::INTEGER[]) AS ordered(id, weight)
		WHERE roles.id = ordered.id
		`,
		pq.StringArray(roleIDs),
		pq.Int64Array(weights),
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) SetUserRoles(ctx context.Context, userID string, roleIDs []string) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	current, err := list[RoleBinding](
		ctx,
		tx,
		`SELECT * FROM role_bindings WHERE user_id = $1 FOR UPDATE`,
		userID,
	)

	if err != nil {
		return false, err
	}

	changed := false
	for _, binding := range current {
		if slices.Contains(roleIDs, binding.RoleID) {
			continue
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM role_bindings WHERE id = $1`, binding.ID); err != nil {
			return false, err
		}

		changed = true
	}

	for _, roleID := range roleIDs {
		exists := slices.ContainsFunc(current, func(binding RoleBinding) bool {
			return binding.RoleID == roleID
		})

		if exists {
			continue
		}

		binding := &RoleBinding{
			ID:        CreateID(RoleBindingKey),
			RoleID:    roleID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}

		if err := db.insertRoleBinding(ctx, tx, binding); err != nil {
			return false, err
		}

		changed = true
	}

	if !changed {
		return false, nil
	}

	if err := db.revokeUser(ctx, tx, userID, RevocationReasonRolesChanged); err != nil {
		return false, err
	}

	return true, tx.Commit()
}
