package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
)

func revocationRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminTokensRead), middleware.RequirePermission(permission.PermissionAdminReadTokens), getTokenRevocations)
	router.Post("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminTokensWrite), middleware.RequirePermission(permission.PermissionAdminRevokeTokens), createTokenRevocation)
}

func getTokenRevocations(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	page, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(star.Database.GetTokenRevocations(ctx, page, limit))
}

func createTokenRevocation(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	var data types.CreateRevocationRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	if data.Kind == string(database.RevocationKindGlobal) {
		err = star.Database.RevokeAllTokens(ctx, database.RevocationReasonManual)
	} else if user := star.Database.GetUserByID(ctx, *data.User); user == nil {
		return fiber.NewError(404, "User not found")
	} else {
		err = star.Database.RevokeUserTokens(ctx, user.ID, database.RevocationReasonManual)
	}

	if err != nil {
		return err
	}

	slog.Info("access tokens revoked", "kind", data.Kind, "user", data.User, "by", requesterID)

	cache.UpdateRevocations(ctx, star)
	return getTokenRevocations(ctx)
}
