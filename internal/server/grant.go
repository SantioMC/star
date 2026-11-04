package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
)

func grantRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserAppsRead), getUserGrants)
	router.Delete(
		"/:application",
		middleware.RequireAuth,
		middleware.RequireScope(openid.ScopeUserAppsRevoke),
		middleware.RequirePermission(permission.PermissionRevokeAuthorization),
		revokeUserGrant,
	)
}

func getUserGrants(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	grants := star.Database.GetUserApplicationGrants(ctx, userID)

	scopes := make([]string, 0, len(grants))
	for _, grant := range grants {
		scopes = append(scopes, grant.Scopes...)
	}

	descriptions := openid.DescribeScopes(ctx, star.Database, scopes)
	return ctx.JSON(transformer.TransformGrants(grants, descriptions))
}

func revokeUserGrant(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	grant := star.Database.GetApplicationGrant(ctx, userID, ctx.Params("application"))
	if grant == nil {
		return fiber.NewError(404, "Application not found")
	}

	if err := star.Database.RevokeApplicationGrant(ctx, userID, grant.ApplicationID); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.GrantRevoke, audit.On(audit.TargetApplication, grant.ApplicationID, ""))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
