package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/util"
	"github.com/tubblelive/star/star/verifier"
	"gorm.io/datatypes"
)

func RoleRouter(r fiber.Router) {
	r.Get("/", getRoles)
	r.Post("/", createRole)

	r.Get("/:role", getRole)
	r.Patch("/:role", patchRole)
	r.Delete("/:role", deleteRole)
}

func getRoles(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadRoles); err != nil {
		return err
	}

	roles, err := service.GetAllRoles()
	if err != nil {
		return err
	}

	return ctx.JSON(roles)
}

func createRole(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminCreateRoles); err != nil {
		return err
	}

	var body CreateRoleRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	name, err := verifier.VerifyRoleName(body.Name)
	if err != nil {
		return helper.NewHttpError(400, err.Error())
	}

	permissions, err := verifyPermissions(body.Permissions)
	if err != nil {
		return err
	}

	role, err := service.CreateRole(name, permissions)
	if err != nil {
		return err
	}

	return ctx.JSON(role)
}

func getRole(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadRoles); err != nil {
		return err
	}

	role, err := service.GetRoleByID(ctx.Params("role"))
	if err != nil {
		return err
	}

	if role == nil {
		return helper.NewHttpError(404, "role not found")
	}

	return ctx.JSON(role)
}

func patchRole(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyRoles); err != nil {
		return err
	}

	role, err := service.GetRoleByID(ctx.Params("role"))
	if err != nil {
		return err
	}

	if role == nil {
		return helper.NewHttpError(404, "role not found")
	}

	var body PatchRoleRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Name != nil {
		name, err := verifier.VerifyRoleName(*body.Name)
		if err != nil {
			return helper.NewHttpError(400, err.Error())
		}

		role.Name = name
		service.RenameRole(role.ID, name)
	}

	if body.Permissions != nil {
		permissions, err := verifyPermissions(body.Permissions)
		if err != nil {
			return helper.NewHttpError(400, err.Error())
		}

		role.Permissions = datatypes.NewJSONType(*permissions)
		service.SetPermissions(role.ID, *permissions)
	}

	return ctx.JSON(role)
}

func deleteRole(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminDeleteRoles); err != nil {
		return err
	}

	role, err := service.GetRoleByID(ctx.Params("role"))
	if err != nil {
		return err
	}

	if role == nil {
		return helper.NewHttpError(404, "role not found")
	}

	if err := service.DeleteRole(role.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{ "success": true })
}

func verifyPermissions(permissions *[]string) (*[]string, error) {
	var sanitizedPermissions []string

	if permissions != nil {
		for _, permission := range util.Unique(*permissions) {
			sanitized, err := verifier.VerifyPermission(permission)
			if err != nil {
				return nil, helper.NewHttpError(400, err.Error())
			}

			sanitizedPermissions = append(sanitizedPermissions, sanitized)
		}
	}

	return &sanitizedPermissions, nil
}

type CreateRoleRequest struct {
	Name        string    `json:"name"`
	Permissions *[]string `json:"scopes"`
}

type PatchRoleRequest struct {
	Name        *string   `json:"name"`
	Permissions *[]string `json:"scopes"`
}
