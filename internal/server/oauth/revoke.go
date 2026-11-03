package oauth

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

func Revoke(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	form, ok, err := parseForm(ctx)
	if !ok {
		return err
	}

	application, ok, err := authenticateClient(ctx, form)
	if !ok {
		return err
	}

	if application == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing client_id, clients must identify themselves")
	}

	value := form.Get("token")
	if value == "" {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing token")
	}

	userID, familyID, ok := revocationTarget(ctx, star, application, value)
	if !ok {
		return ctx.SendStatus(200)
	}

	if err := star.Database.RevokeSessionFamily(ctx, userID, familyID, database.RevocationReasonSignedOut); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	return ctx.SendStatus(200)
}

func revocationTarget(ctx fiber.Ctx, star *state.StarServer, application *database.Application, value string) (string, string, bool) {
	if strings.Count(value, ".") == 2 {
		jwt, err := token.Verify(star, value)
		if err != nil {
			return "", "", false
		}

		if jwt.ClientID != application.ClientID {
			slog.Warn("client attempted to revoke an access token issued to another client", "client", application.ID, "session", jwt.SessionID)
			return "", "", false
		}

		return jwt.Subject, jwt.SessionID, true
	}

	session := star.Database.GetSessionByToken(ctx, value)
	if session == nil {
		return "", "", false
	}

	if session.ApplicationID != application.ID {
		slog.Warn("client attempted to revoke a token belonging to another client", "client", application.ID, "session", session.FamilyID)
		return "", "", false
	}

	return session.UserID, session.FamilyID, true
}
