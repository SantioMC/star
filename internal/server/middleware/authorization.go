package middleware

import (
	"errors"
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/token"
)

func RequirePermission(permission permission.PermissionDefinition) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		value := ctx.Locals(RequestAuthJWTKey)
		if value == nil {
			return errors.New("missing auth jwt key in context")
		}

		jwt := value.(*token.AccessTokenJWT)
		if jwt == nil {
			return errors.New("no jwt found in request, there is likely no RequireAuth before the RequirePermission middleware")
		}

		if slices.Contains(jwt.Permissions, "*") {
			return ctx.Next()
		}

		if !slices.Contains(jwt.Permissions, permission.Name) {
			return fiber.NewError(403, "You are not authorized to perform this action")
		}

		return ctx.Next()
	}
}
