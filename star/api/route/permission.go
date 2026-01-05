package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func PermissionRouter(r fiber.Router) {
	r.Get("/", getAllPermissions)
}

func getAllPermissions(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return helper.NewHttpError(404, "app not found")
	}

	permissions := service.GetPermissions(app.ID)
	return ctx.JSON(permissions)
}