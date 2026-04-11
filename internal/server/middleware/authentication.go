package middleware

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/token"
)

var RequestAuthJWTKey = "request.auth.jwt"

func RequireAuth(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	auth := ctx.Get(fiber.HeaderAuthorization)
	if auth == "" {
		return fiber.NewError(401, "Unauthorized")
	}

	segments := strings.Split(auth, " ")
	if len(segments) != 2 {
		return fiber.NewError(401, "Unauthorized")
	}

	if strings.ToLower(segments[0]) != "bearer" {
		return fiber.NewError(401, "Unauthorized")
	}

	value := segments[1]
	if strings.Count(value, ".") != 2 {
		return fiber.NewError(401, "Unauthorized")
	}

	kid, err := getKeyId(value)
	if err != nil {
		return err
	}

	var signingKey *database.SigningKey = nil
	for _, key := range *star.SigningKeys {
		if key.ID == kid {
			signingKey = &key
			break
		}
	}

	if signingKey == nil {
		return fiber.NewError(401, "Unauthorized")
	}

	publicKey, err := signingKey.GetPublicKey()
	if err != nil {
		return err
	}

	jwt, err := token.VerifyJWT(value, publicKey)
	if err != nil || jwt == nil {
		slog.Warn("failed to verify user jwt", "error", err)
		return fiber.NewError(401, "Unauthorized")
	}

	ctx.Locals(RequestAuthJWTKey, jwt)
	return ctx.Next()
}

func getKeyId(jwt string) (string, error) {
	segments := strings.Split(jwt, ".")
	if len(segments) != 3 {
		return "", fiber.NewError(401, "Unauthorized")
	}

	header, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		slog.Warn("failed to decode jwt header", "error", err)
		return "", fiber.NewError(401, "Unauthorized")
	}

	var headerMap map[string]interface{}
	if err := json.Unmarshal(header, &headerMap); err != nil {
		slog.Warn("failed to unmarshal jwt header", "error", err, "header", string(header))
		return "", fiber.NewError(401, "Unauthorized")
	}

	kid, ok := headerMap["kid"].(string)
	if !ok {
		slog.Warn("failed to get kid from jwt header", "header", string(header))
		return "", fiber.NewError(401, "Unauthorized")
	}

	return kid, nil
}
