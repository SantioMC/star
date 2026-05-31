package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/middleware"
)

func mfaRouter(router fiber.Router) {
	router.Post("/totp/enroll", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionEnrollMFA), enrollTotp)
}

func enrollTotp(ctx fiber.Ctx) error {
	return nil
}
