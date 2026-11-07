package server

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/federation"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func connectionRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserAccountRead), getConnections)
	router.Post("/:provider", middleware.RateLimit(10), middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserAccountWrite), middleware.RequirePermission(permission.PermissionManageConnections), startConnection)
	router.Delete("/:provider", middleware.RateLimit(10), middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserAccountWrite), middleware.RequirePermission(permission.PermissionManageConnections), deleteConnection)
}

func getConnections(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	providers := star.Database.GetProviders(ctx)
	accounts := star.Database.GetAccountsByUserID(ctx, userID)
	includeEmail := request.HasScope(ctx, openid.ScopeUserEmailRead)

	return ctx.JSON(transformer.TransformConnections(providers, accounts, includeEmail))
}

func startConnection(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	provider := getEnabledProvider(ctx, star, ctx.Params("provider"))
	if provider == nil {
		return fiber.NewError(404, "This sign in provider is no longer available")
	}

	if star.Database.GetUserProviderAccount(ctx, jwt.Subject, provider.ID) != nil {
		return federation.ErrProviderLinked
	}

	target, err := beginProviderAuthorization(ctx, star, provider, jwt.Subject, jwt.SessionID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"url": target,
	})
}

func completeConnection(ctx fiber.Ctx, star *state.StarServer, loginState *types.ProviderLoginState) error {
	account, err := connectProvider(ctx, star, loginState)
	if err != nil {
		if known, ok := errors.AsType[*fiber.Error](err); ok {
			return redirectToProfile(ctx, star, "connection_error", known.Message)
		}

		slog.Error("provider connection failed", "provider", loginState.Provider, "err", err)
		return redirectToProfile(ctx, star, "connection_error", "Unable to connect this provider, please try again later")
	}

	audit.Log(
		ctx,
		audit.ProviderAssociate,
		audit.ByID(account.UserID),
		audit.On(audit.TargetProvider, loginState.Provider, ""),
		audit.Private("external_id", *account.ExternalID),
	)

	return redirectToProfile(ctx, star, "connected", loginState.Provider)
}

func connectProvider(ctx fiber.Ctx, star *state.StarServer, loginState *types.ProviderLoginState) (*database.Account, error) {
	session := star.Database.GetSessionByFamily(ctx, loginState.FamilyID)
	if session == nil || session.UserID != loginState.UserID {
		return nil, fiber.NewError(401, "Your session expired, please sign in and try again")
	}

	user := star.Database.GetUserByID(ctx, loginState.UserID)
	if user == nil {
		return nil, fiber.NewError(404, "User not found")
	}

	provider, identity, err := fetchProviderIdentity(ctx, star, loginState, callbackValue(ctx, "code"))
	if err != nil {
		return nil, err
	}

	return federation.LinkAccount(ctx, star, provider, user, identity)
}

func deleteConnection(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	account := star.Database.GetUserProviderAccount(ctx, jwt.Subject, ctx.Params("provider"))
	if account == nil {
		return fiber.NewError(404, "This provider is not connected to your account")
	}

	removed, err := star.Database.RemoveProviderAccount(ctx, account, jwt.SessionID)
	if err != nil {
		return err
	} else if !removed {
		return fiber.NewError(400, "You must add another form of authentication before removing this connection")
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(
		ctx,
		audit.ProviderDisassociate,
		audit.On(audit.TargetProvider, *account.ProviderID, ""),
		audit.Private("external_id", *account.ExternalID),
	)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
