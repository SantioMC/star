package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/token"
)

var RequestAuthJWTKey = "request.auth.jwt"

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

	if !slices.Contains(jwt.Audience, star.DefaultApp.ID) {
		return unauthorized(ctx, "This token is not permitted to access this application")
	}

	ctx.Locals(RequestAuthJWTKey, jwt)
	return ctx.Next()
}

func RequireOAuth(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := verifyBearer(ctx)
	if err != nil {
		return err
	}

	for _, audience := range jwt.Audience {
		if audience == star.DefaultApp.ID {
			continue
		}

		if star.Database.GetApplicationByID(ctx, audience) == nil {
			return unauthorized(ctx, "The token audience is not a registered application")
		}
	}

	ctx.Locals(RequestAuthJWTKey, jwt)
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

	value := segments[1]
	if strings.Count(value, ".") != 2 {
		return nil, unauthorized(ctx, "The access token is malformed")
	}

	kid, err := getKeyID(ctx, value)
	if err != nil {
		return nil, err
	}

	var signingKey *database.SigningKey = nil
	for _, key := range *star.SigningKeys {
		if key.ID == kid {
			signingKey = &key
			break
		}
	}

	if signingKey == nil {
		return nil, unauthorized(ctx, "The access token was signed by an unknown key")
	}

	publicKey, err := signingKey.GetPublicKey()
	if err != nil {
		return nil, err
	}

	jwt, err := token.VerifyJWT(star, value, publicKey)
	if err != nil || jwt == nil {
		slog.Warn("failed to verify user jwt", "error", err)
		return nil, unauthorized(ctx, "The access token is expired or invalid")
	}

	return jwt, nil
}

func getKeyID(ctx fiber.Ctx, jwt string) (string, error) {
	segments := strings.Split(jwt, ".")
	if len(segments) != 3 {
		return "", unauthorized(ctx, "The access token is malformed")
	}

	header, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		slog.Warn("failed to decode jwt header", "error", err)
		return "", unauthorized(ctx, "The access token is malformed")
	}

	var headerMap map[string]any
	if err := json.Unmarshal(header, &headerMap); err != nil {
		slog.Warn("failed to unmarshal jwt header", "error", err, "header", string(header))
		return "", unauthorized(ctx, "The access token is malformed")
	}

	kid, ok := headerMap["kid"].(string)
	if !ok {
		slog.Warn("failed to get kid from jwt header", "header", string(header))
		return "", unauthorized(ctx, "The access token is missing a key id")
	}

	return kid, nil
}
