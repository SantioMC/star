package server

import (
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
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

	router.Get("/", readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getRoles)
	router.Post("/", writeRoles, middleware.RequirePermission(permission.PermissionAdminCreateRoles), createRole)

	router.Get("/permissions", readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getAvailablePermissions)

	router.Get("/:role", readRoles, middleware.RequirePermission(permission.PermissionAdminReadRoles), getRole)
	router.Patch("/:role", writeRoles, middleware.RequirePermission(permission.PermissionAdminWriteRoles), updateRole)
	router.Delete("/:role", writeRoles, middleware.RequirePermission(permission.PermissionAdminDeleteRoles), deleteRole)
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

	if err := star.Database.CreateRole(ctx, role); err != nil {
		return err
	}

	if err := bindPermissions(ctx, star, role.ID, permissions); err != nil {
		return err
	}

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

	if role.Internal {
		return fiber.NewError(403, "You are not permitted to modify this role")
	}

	if request.Name != nil && *request.Name != role.Name {
		if star.Database.GetRoleByName(ctx, *request.Name) != nil {
			return fiber.NewError(400, "A role with this name already exists")
		}

		role.Name = *request.Name
	}

	if request.Description != nil {
		role.Description = *request.Description
	}

	if err := star.Database.UpdateRole(ctx, role); err != nil {
		return err
	}

	if request.Permissions != nil {
		permissions, err := resolvePermissions(ctx, star, *request.Permissions)
		if err != nil {
			return err
		}

		if err := star.Database.DeletePermissionBindingsByRoleID(ctx, role.ID); err != nil {
			return err
		}

		if err := bindPermissions(ctx, star, role.ID, permissions); err != nil {
			return err
		}
	}

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

	if err := star.Database.DeleteRoleByID(ctx, role.ID); err != nil {
		return err
	}

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

func bindPermissions(ctx fiber.Ctx, star *state.StarServer, roleID string, permissions []database.Permission) error {
	for _, resolved := range permissions {
		binding := &database.PermissionBinding{
			ID:           database.CreateID(database.PermissionBindingKey),
			PermissionID: resolved.ID,
			RoleID:       roleID,
			CreatedAt:    time.Now(),
		}

		if err := star.Database.CreatePermissionBinding(ctx, binding); err != nil {
			return err
		}
	}

	return nil
}
