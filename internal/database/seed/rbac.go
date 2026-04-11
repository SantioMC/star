package seed

import (
	"context"
	"sync"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/state"
)

func seedRBAC(ctx context.Context, star *state.StarServer) error {
	role := star.Database.GetRoleByName(ctx, "superadmin")
	if role == nil {
		role = &database.Role{
			ID:          database.CreateID("rol"),
			Name:        "superadmin",
			Description: "The default super admin role",
			CreatedAt:   time.Now(),
		}

		if err := star.Database.CreateRole(ctx, role); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(permission.BuiltinPermissions))

	errors := make(chan error, len(permission.BuiltinPermissions))

	for _, permission := range permission.BuiltinPermissions {
		go func() {
			defer wg.Done()

			existing := star.Database.GetPermissionByName(ctx, permission.Name)
			if existing != nil {
				return
			}

			permission := &database.Permission{
				ID:          database.CreateID("perm"),
				Name:        permission.Name,
				Description: permission.Description,
				CreatedAt:   time.Now(),
			}

			if err := star.Database.CreatePermission(ctx, permission); err != nil {
				errors <- err
				return
			}

			binding := &database.PermissionBinding{
				ID:           database.CreateID("pb"),
				PermissionID: permission.ID,
				RoleID:       role.ID,
				CreatedAt:    time.Now(),
			}

			if err := star.Database.CreatePermissionBinding(ctx, binding); err != nil {
				errors <- err
				return
			}
		}()
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
