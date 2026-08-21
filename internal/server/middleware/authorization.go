package middleware

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/token"
)

func requestJWT(ctx fiber.Ctx) (*token.AccessTokenJWT, error) {
	value := ctx.Locals(RequestAuthJWTKey)
	if value == nil {
		return nil, errors.New("no jwt found in request, there is likely no authentication middleware before this one")
	}

	jwt, ok := value.(*token.AccessTokenJWT)
	if !ok || jwt == nil {
		return nil, errors.New("the request auth jwt key does not hold a jwt")
	}

	return jwt, nil
}

func insufficientScope(ctx fiber.Ctx, scope string) error {
	ctx.Set(
		fiber.HeaderWWWAuthenticate,
		fmt.Sprintf(`Bearer error="insufficient_scope", error_description=%q, scope=%q`, "The "+scope+" scope is required", scope),
	)

	return fiber.NewError(403, "This token has not been granted the "+scope+" scope")
}

func RequirePermission(permission permission.PermissionDefinition) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		jwt, err := requestJWT(ctx)
		if err != nil {
			return err
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

func RequireScope(scope string) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		star := common.GetState(ctx)

		jwt, err := requestJWT(ctx)
		if err != nil {
			return err
		}

		firstParty := slices.Contains(jwt.Audience, star.DefaultApp.ID)
		if openid.IsStarScope(scope) && !firstParty {
			return insufficientScope(ctx, scope)
		}

		granted := openid.EffectiveScopes(strings.Fields(jwt.Scope), firstParty)
		if !openid.HasScope(granted, scope) {
			return insufficientScope(ctx, scope)
		}

		return ctx.Next()
	}
}
