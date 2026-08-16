package oauth

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
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

	if strings.Count(value, ".") == 2 {
		return writeTokenError(ctx, 400, Oauth2ErrorUnsupportedTokenType, "Access tokens cannot be revoked, they remain valid until they expire")
	}

	session := star.Database.GetSessionByToken(ctx, value)
	if session == nil {
		return ctx.SendStatus(200)
	}

	if session.ApplicationID != application.ID {
		slog.Warn("client attempted to revoke a token belonging to another client", "client", application.ID, "session", session.ID)
		return ctx.SendStatus(200)
	}

	if err := star.Database.DeleteSessionByID(ctx, session.ID); err != nil {
		return err
	}

	return ctx.SendStatus(200)
}
