package server

import (
	"errors"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func roleRouter(router fiber.Router) {
	readRoles := middleware.RequireScope(openid.ScopeAdminRolesRead)
	writeRoles := middleware.RequireScope(openid.ScopeAdminRolesWrite)

	router.Get("/", middleware.RequireAuth, readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getRoles)
	router.Post("/", middleware.RequireAuth, writeRoles, middleware.RequirePermission(permission.PermissionAdminCreateRoles), createRole)

	router.Get("/permissions", middleware.RequireAuth, readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getAvailablePermissions)

	router.Post("/order", middleware.RequireAuth, writeRoles, middleware.RequirePermission(permission.PermissionAdminWriteRoles), reorderRoles)

	router.Get("/:role", middleware.RequireAuth, readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getRole)
	router.Patch("/:role", middleware.RequireAuth, writeRoles, middleware.RequirePermission(permission.PermissionAdminWriteRoles), updateRole)
	router.Delete("/:role", middleware.RequireAuth, writeRoles, middleware.RequirePermission(permission.PermissionAdminDeleteRoles), deleteRole)
}

func getAvailablePermissions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	return ctx.JSON(star.Database.GetPermissions(ctx))
}

func getRoles(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	roles := star.Database.GetRoles(ctx)
	transformed := make([]transformer.AdminRole, len(roles))

	for idx, role := range roles {
		permissions := star.Database.GetRolePermissions(ctx, role.ID)
		transformed[idx] = transformer.TransformRole(role, permissions, reach)
	}

	return ctx.JSON(transformed)
}

func requireRoleReach(ctx fiber.Ctx, role *database.Role) (int, error) {
	reach, err := request.GetWeight(ctx)
	if err != nil {
		return 0, err
	}

	if role.Weight > reach {
		return 0, fiber.NewError(403, "You are not permitted to manage a role higher than your own")
	}

	return reach, nil
}

func reorderRoles(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.ReorderRolesRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	err = star.Database.ReorderRoles(ctx, data.Roles, reach)
	if errors.Is(err, database.ErrRoleOrderOutdated) {
		return fiber.NewError(409, "The roles have changed, refresh the page and try again")
	} else if errors.Is(err, database.ErrRoleOrderForbidden) {
		return fiber.NewError(403, "You are not permitted to move a role at or above your own")
	} else if err != nil {
		return err
	}

	audit.Log(ctx, audit.RoleReorder, audit.With("roles", data.Roles))

	return getRoles(ctx)
}

func getRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	role := star.Database.GetRoleByID(ctx, ctx.Params("role"))
	if role == nil {
		return fiber.NewError(404, "Role not found")
	}

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	permissions := star.Database.GetRolePermissions(ctx, role.ID)
	return ctx.JSON(transformer.TransformRole(*role, permissions, reach))
}

func createRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.CreateRoleRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	if star.Database.GetRoleByName(ctx, data.Name) != nil {
		return fiber.NewError(400, "A role with this name already exists")
	}

	permissions, err := resolvePermissions(ctx, star, data.Permissions)
	if err != nil {
		return err
	}

	role := &database.Role{
		ID:          database.CreateID(database.RoleKey),
		Name:        data.Name,
		Description: data.Description,
		Internal:    false,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreateRoleWithPermissions(ctx, role, permissions); errors.Is(err, database.ErrDuplicate) {
		return fiber.NewError(409, "Another role was created at the same time, please try again")
	} else if err != nil {
		return err
	}

	audit.Log(ctx, audit.RoleCreate, audit.On(audit.TargetRole, role.ID, role.Name), audit.With("permissions", data.Permissions))

	return ctx.JSON(transformer.TransformRole(*role, permissions, reach))
}

func updateRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.UpdateRoleRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	role := star.Database.GetRoleByID(ctx, ctx.Params("role"))
	if role == nil {
		return fiber.NewError(404, "Role not found")
	}

	reach, err := requireRoleReach(ctx, role)
	if err != nil {
		return err
	}

	before := *role
	nameChanged := data.Name != nil && *data.Name != role.Name
	descriptionChanged := data.Description != nil && *data.Description != role.Description

	if role.IsDefault() && (nameChanged || descriptionChanged) {
		return fiber.NewError(403, "Only the permissions of the default role can be modified")
	} else if role.Internal && !role.IsDefault() {
		return fiber.NewError(403, "You are not permitted to modify this role")
	}

	if nameChanged {
		if star.Database.GetRoleByName(ctx, *data.Name) != nil {
			return fiber.NewError(400, "A role with this name already exists")
		}

		role.Name = *data.Name
	}

	if data.Description != nil {
		role.Description = *data.Description
	}

	if data.Permissions == nil {
		if err := star.Database.UpdateRole(ctx, role); err != nil {
			return err
		}
	} else {
		permissions, err := resolvePermissions(ctx, star, *data.Permissions)
		if err != nil {
			return err
		}

		if err := star.Database.UpdateRoleWithPermissions(ctx, role, permissions); err != nil {
			return err
		}

		cache.UpdateRevocations(ctx, star)
	}

	options := []audit.Option{audit.On(audit.TargetRole, role.ID, role.Name), audit.Changes(before, *role)}
	if data.Permissions != nil {
		options = append(options, audit.With("permissions", *data.Permissions))
	}

	audit.Log(ctx, audit.RoleUpdate, options...)

	permissions := star.Database.GetRolePermissions(ctx, role.ID)
	return ctx.JSON(transformer.TransformRole(*role, permissions, reach))
}

func deleteRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	role := star.Database.GetRoleByID(ctx, ctx.Params("role"))
	if role == nil {
		return fiber.NewError(404, "Role not found")
	}

	if role.Internal {
		return fiber.NewError(403, "You are not permitted to delete this role")
	}

	if _, err := requireRoleReach(ctx, role); err != nil {
		return err
	}

	if err := star.Database.DeleteRoleWithRevocation(ctx, role); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.RoleDelete, audit.On(audit.TargetRole, role.ID, role.Name))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func resolvePermissions(ctx fiber.Ctx, star *state.StarServer, names []string) ([]database.Permission, error) {
	permissions := make([]database.Permission, 0, len(names))

	for _, name := range slices.Compact(slices.Sorted(slices.Values(names))) {
		resolved := star.Database.GetPermissionByName(ctx, name)
		if resolved == nil {
			return nil, fiber.NewError(400, "Unknown permission: "+name)
		}

		permissions = append(permissions, *resolved)
	}

	return permissions, nil
}
