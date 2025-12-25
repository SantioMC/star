package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func RoleRouter(r fiber.Router) {
	r.Get("/", getRoles)
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