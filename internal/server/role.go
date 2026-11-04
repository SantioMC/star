package server

import (
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

	roles := star.Database.GetRoles(ctx)
	transformed := make([]transformer.AdminRole, len(roles))

	for idx, role := range roles {
		permissions := star.Database.GetRolePermissions(ctx, role.ID)
		transformed[idx] = transformer.TransformRole(role, permissions)
	}

	return ctx.JSON(transformed)
}

func getRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	role := star.Database.GetRoleByID(ctx, ctx.Params("role"))
	if role == nil {
		return fiber.NewError(404, "Role not found")
	}

	permissions := star.Database.GetRolePermissions(ctx, role.ID)
	return ctx.JSON(transformer.TransformRole(*role, permissions))
}

func createRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.CreateRoleRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if star.Database.GetRoleByName(ctx, request.Name) != nil {
		return fiber.NewError(400, "A role with this name already exists")
	}

	permissions, err := resolvePermissions(ctx, star, request.Permissions)
	if err != nil {
		return err
	}

	role := &database.Role{
		ID:          database.CreateID(database.RoleKey),
		Name:        request.Name,
		Description: request.Description,
		Internal:    false,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreateRoleWithPermissions(ctx, role, permissions); err != nil {
		return err
	}

	audit.Log(ctx, audit.RoleCreate, audit.On(audit.TargetRole, role.ID, role.Name), audit.With("permissions", request.Permissions))

	return ctx.JSON(transformer.TransformRole(*role, permissions))
}

func updateRole(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.UpdateRoleRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	role := star.Database.GetRoleByID(ctx, ctx.Params("role"))
	if role == nil {
		return fiber.NewError(404, "Role not found")
	}

	before := *role
	nameChanged := request.Name != nil && *request.Name != role.Name
	descriptionChanged := request.Description != nil && *request.Description != role.Description

	if role.IsDefault() && (nameChanged || descriptionChanged) {
		return fiber.NewError(403, "Only the permissions of the default role can be modified")
	} else if role.Internal && !role.IsDefault() {
		return fiber.NewError(403, "You are not permitted to modify this role")
	}

	if nameChanged {
		if star.Database.GetRoleByName(ctx, *request.Name) != nil {
			return fiber.NewError(400, "A role with this name already exists")
		}

		role.Name = *request.Name
	}

	if request.Description != nil {
		role.Description = *request.Description
	}

	if request.Permissions == nil {
		if err := star.Database.UpdateRole(ctx, role); err != nil {
			return err
		}
	} else {
		permissions, err := resolvePermissions(ctx, star, *request.Permissions)
		if err != nil {
			return err
		}

		if err := star.Database.UpdateRoleWithPermissions(ctx, role, permissions); err != nil {
			return err
		}

		cache.UpdateRevocations(ctx, star)
	}

	options := []audit.Option{audit.On(audit.TargetRole, role.ID, role.Name), audit.Changes(before, *role)}
	if request.Permissions != nil {
		options = append(options, audit.With("permissions", *request.Permissions))
	}

	audit.Log(ctx, audit.RoleUpdate, options...)

	permissions := star.Database.GetRolePermissions(ctx, role.ID)
	return ctx.JSON(transformer.TransformRole(*role, permissions))
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
