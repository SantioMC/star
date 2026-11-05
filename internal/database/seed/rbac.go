package seed

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/state"
)

func seedRBAC(ctx context.Context, star *state.StarServer) error {
	superadmin, superadminCreated, err := seedRole(ctx, star, database.SuperadminRoleName, "The default super admin role", database.SuperadminRoleWeight)
	if err != nil {
		return err
	}

	defaultRole, defaultCreated, err := seedRole(ctx, star, database.DefaultRoleName, "Permissions granted to every user", database.DefaultRoleWeight)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(permission.BuiltinPermissions))

	for _, definition := range permission.BuiltinPermissions {
		wg.Go(func() {
			permission, created, err := seedPermission(ctx, star, definition)
			if err != nil {
				errors <- err
				return
			}

			if superadmin != nil && (created || superadminCreated) {
				if err := bindPermission(ctx, star, superadmin, permission); err != nil {
					errors <- err
					return
				}
			}

			if defaultRole != nil && definition.Default && (created || defaultCreated) {
				if err := bindPermission(ctx, star, defaultRole, permission); err != nil {
					errors <- err
				}
			}
		})
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func seedRole(ctx context.Context, star *state.StarServer, name string, description string, weight int) (*database.Role, bool, error) {
	existing := star.Database.GetRoleByName(ctx, name)
	if existing != nil && !existing.Internal {
		slog.Warn("a non built-in role is using a reserved name, it will not be managed by star", "role", existing.ID, "name", name)
		return nil, false, nil
	} else if existing != nil {
		return existing, false, nil
	}

	role := &database.Role{
		ID:          database.CreateID(database.RoleKey),
		Name:        name,
		Description: description,
		Internal:    true,
		Weight:      weight,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreateRole(ctx, role); err != nil {
		return nil, false, err
	}

	return role, true, nil
}

func seedPermission(ctx context.Context, star *state.StarServer, definition permission.PermissionDefinition) (*database.Permission, bool, error) {
	existing := star.Database.GetPermissionByName(ctx, definition.Name)
	if existing != nil {
		return existing, false, nil
	}

	permission := &database.Permission{
		ID:          database.CreateID(database.PermissionKey),
		Name:        definition.Name,
		Description: definition.Description,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreatePermission(ctx, permission); err != nil {
		return nil, false, err
	}

	return permission, true, nil
}

func bindPermission(ctx context.Context, star *state.StarServer, role *database.Role, permission *database.Permission) error {
	return star.Database.CreatePermissionBinding(ctx, &database.PermissionBinding{
		ID:           database.CreateID(database.PermissionBindingKey),
		PermissionID: permission.ID,
		RoleID:       role.ID,
		CreatedAt:    time.Now(),
	})
}
