package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/token"
)

func unauthorized(ctx fiber.Ctx, description string) error {
	challenge := `Bearer realm="star"`
	if description != "" {
		challenge += fmt.Sprintf(`, error="invalid_token", error_description=%q`, description)
	}

	ctx.Set(fiber.HeaderWWWAuthenticate, challenge)
	return fiber.NewError(401, "Unauthorized")
}

func RequireAuth(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := verifyBearer(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(jwt.Audience, star.DefaultApp.Load().ID) {
		return unauthorized(ctx, "This token is not permitted to access this application")
	}

	ctx.Locals(common.RequestAuthJWTKey, jwt)
	return ctx.Next()
}

func OptionalAuth(ctx fiber.Ctx) error {
	if ctx.Get(fiber.HeaderAuthorization) == "" {
		return ctx.Next()
	}

	return RequireAuth(ctx)
}

func RequireOAuth(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := verifyBearer(ctx)
	if err != nil {
		return err
	}

	defaultApp := star.DefaultApp.Load()
	for _, audience := range jwt.Audience {
		if audience == defaultApp.ID {
			continue
		}

		if star.Database.GetApplicationByID(ctx, audience) == nil {
			return unauthorized(ctx, "The token audience is not a registered application")
		}
	}

	ctx.Locals(common.RequestAuthJWTKey, jwt)
	return ctx.Next()
}

func verifyBearer(ctx fiber.Ctx) (*token.AccessTokenJWT, error) {
	star := common.GetState(ctx)

	auth := ctx.Get(fiber.HeaderAuthorization)
	if auth == "" {
		return nil, unauthorized(ctx, "")
	}

	segments := strings.Split(auth, " ")
	if len(segments) != 2 {
		return nil, unauthorized(ctx, "Malformed authorization header")
	}

	if strings.ToLower(segments[0]) != "bearer" {
		return nil, unauthorized(ctx, "Only the Bearer authentication scheme is supported")
	}

	jwt, err := token.Verify(star, segments[1])
	if errors.Is(err, token.ErrMalformedToken) || errors.Is(err, token.ErrMissingKeyID) || errors.Is(err, token.ErrUnknownKey) {
		return nil, unauthorized(ctx, err.Error())
	} else if err != nil {
		slog.Warn("failed to verify user jwt", "error", err)
		return nil, unauthorized(ctx, "The access token is expired or invalid")
	}

	return jwt, nil
}
