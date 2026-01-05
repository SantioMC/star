package helper

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/state/claims"
	"github.com/tubblelive/star/star/state/scopes"
)

func getBearerToken(ctx *fiber.Ctx) (string, error) {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return "", NewHttpError(401, "Unauthorized")
	}

	bearer := strings.Split(authHeader, " ")
	if len(bearer) != 2 || bearer[0] != "Bearer" {
		return "", NewHttpError(401, "Unauthorized")
	}

	return bearer[1], nil
}

func GetUserClaims(ctx *fiber.Ctx) (*claims.UserJWT, error) {
	token, err := getBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	signingKey, err := state.StarState.SigningKey.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	claims, err := crypto.VerifyJWT(token, &signingKey.PublicKey)
	if err != nil {
		slog.Warn("client attempted to provide unverified jwt", "token", token, "err", err)
		return nil, NewHttpError(401, "Unauthorized")
	}

	return claims, nil
}

func GetUserId(ctx *fiber.Ctx) (string, error) {
	claims, err := GetUserClaims(ctx)
	if err != nil {
		return "", err
	}
	
	return claims.Subject, nil
}

func HasPermission(ctx *fiber.Ctx, permission scopes.Scope) bool {
	claims, err := GetUserClaims(ctx)
	if err != nil {
		return false
	}

	for _, scope := range claims.Scopes {
		if scopes.Permits(scope, permission.Node) {
			return true
		}
	}

	return false
}

func EnforcePermission(ctx *fiber.Ctx, permission scopes.Scope) error {
	if !HasPermission(ctx, permission) {
		return NewHttpError(403, "Forbidden")
	}

	return nil
}