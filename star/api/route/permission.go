package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/verifier"
)

func PermissionRouter(r fiber.Router) {
	r.Get("/", getAllPermissions)
	r.Post("/", createPermission)
	r.Delete("/:permission", deletePermission)
}

func getAllPermissions(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	permissions := service.GetPermissions(app.ID)
	return ctx.JSON(permissions)
}

func createPermission(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	var body CreatePermissionRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	sanitized, err := verifier.VerifyPermission(body.Permission)
	if err != nil {
		return helper.NewHttpError(400, err.Error())
	}

	permission, err := service.CreatePermission(app, sanitized, body.Description)
	if err != nil {
		return err
	}

	return ctx.JSON(permission)
}

func deletePermission(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	perm, err := service.GetPermissionByID(ctx.Params("permission"))
	if err != nil {
		return err
	}

	if perm == nil || perm.ApplicationID != app.ID {
		return helper.NewHttpError(404, "permission not found")
	}

	if err := service.DeletePermission(perm.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{ "success": true })
}

type CreatePermissionRequest struct {
	Permission  string `json:"name"`
	Description string `json:"description"`
}
