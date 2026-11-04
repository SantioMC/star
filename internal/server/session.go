package server

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/login"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

func sessionRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserSessionsRead), getUserSessions)
	router.Delete("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserSessionsWrite), middleware.RequirePermission(permission.PermissionDeleteSessions), invalidateUserSessions)

	router.Get("/:session", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserSessionsRead), getUserSession)
	router.Delete("/:session", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserSessionsWrite), middleware.RequirePermission(permission.PermissionDeleteSessions), invalidateUserSession)
}

func getUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	sessions := star.Database.GetUserSessions(ctx, userID)
	return ctx.JSON(transformer.TransformSessions(sessions))
}

func invalidateUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	if err := star.Database.RevokeUserSessions(ctx, userID, database.RevocationReasonSessionsCleared); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.SessionRevokeAll)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func getUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	session := star.Database.GetSessionByFamily(ctx, ctx.Params("session"))
	if session == nil || session.UserID != userID {
		return fiber.NewError(404, "session not found")
	}

	return ctx.JSON(transformer.TransformSession(*session))
}

func invalidateUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	session := star.Database.GetSessionByFamily(ctx, ctx.Params("session"))
	if session == nil || session.UserID != userID {
		return fiber.NewError(404, "session not found")
	}

	if err := star.Database.RevokeSessionFamily(ctx, userID, session.FamilyID, database.RevocationReasonSignedOut); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.SessionRevoke, audit.With("session", session.FamilyID))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func setSessionCookie(ctx fiber.Ctx, star *state.StarServer, session *database.Session) {
	request.SetCookie(ctx, star, fiber.Cookie{
		Name:     "star_session",
		Value:    session.Token,
		Path:     "/api/user/session",
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

func clearSessionCookie(ctx fiber.Ctx, star *state.StarServer) {
	request.SetCookie(ctx, star, fiber.Cookie{
		Name:     "star_session",
		Path:     "/api/user/session",
		MaxAge:   -1,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

func cookieSession(ctx fiber.Ctx, star *state.StarServer) *database.Session {
	value := ctx.Cookies("star_session")
	if value == "" {
		return nil
	}

	current := star.Database.GetSessionByToken(ctx, value)
	if current == nil || current.ApplicationID != star.DefaultApp.Load().ID {
		return nil
	}

	return current
}

func refreshUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	current := cookieSession(ctx, star)
	if current == nil {
		clearSessionCookie(ctx, star)
		return fiber.NewError(401, "You are not signed in")
	}

	user := star.Database.GetUserByID(ctx, current.UserID)
	if user == nil {
		clearSessionCookie(ctx, star)
		return fiber.NewError(401, "You are not signed in")
	}

	defaultApp := star.DefaultApp.Load()

	refreshed, err := session.RefreshSession(ctx, star, current, defaultApp.SessionDuration.Duration())
	if errors.Is(err, session.ErrSessionRotated) || errors.Is(err, session.ErrSessionExpired) || errors.Is(err, session.ErrUserSuspended) {
		clearSessionCookie(ctx, star)
		return fiber.NewError(401, "Your session has expired, please sign in again")
	} else if err != nil {
		return err
	}

	accessToken, err := login.IssueAccessToken(ctx, star, refreshed, user)
	if errors.Is(err, token.ErrSessionRevoked) {
		clearSessionCookie(ctx, star)
		return fiber.NewError(401, "Your session has expired, please sign in again")
	} else if err != nil {
		return err
	}

	setSessionCookie(ctx, star, refreshed)

	ctx.Set(fiber.HeaderCacheControl, "no-store")
	return ctx.JSON(types.SessionRefreshResponse{
		AccessToken: string(accessToken),
		ExpiresIn:   int64(defaultApp.TokenDuration.Duration().Seconds()),
	})
}

func endUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	clearSessionCookie(ctx, star)

	current := cookieSession(ctx, star)
	if current == nil {
		return ctx.SendStatus(204)
	}

	if err := star.Database.RevokeSessionFamily(ctx, current.UserID, current.FamilyID, database.RevocationReasonSignedOut); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.SessionEnd, audit.ByID(current.UserID), audit.With("session", current.FamilyID))

	return ctx.SendStatus(204)
}
