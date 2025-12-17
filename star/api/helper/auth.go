package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/state/claims"
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

func EnforcePermission(ctx *fiber.Ctx, permission string) error {
	claims, err := GetUserClaims(ctx)
	if err != nil {
		return err
	}

	// todo: add permissions
	for _, role := range claims.Roles {
		if role.SuperRole {
			return nil
		}
	}

	return NewHttpError(403, "Forbidden")
}