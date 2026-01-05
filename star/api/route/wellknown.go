package route

import (
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/state"
)

func WellKnownRouter(r fiber.Router) {
	r.Get("security.txt", getWellKnownSecurity)
	r.Get("jwks.json", getJWKSKeys)
	r.Get("change-password", getChangePassword)
}

func getWellKnownSecurity(ctx *fiber.Ctx) error {
	expiresAt := star.StartedAt.Add(time.Hour * 24 * 365)
	
	// todo: make configurable
	ctx.SendString(strings.TrimSpace(
	"Contact: mailto:business@santio.me\n" +
	"Expires: " + expiresAt.UTC().Format(time.RFC3339Nano) + "\n" +
	"Preferred-Languages: en",
	))
	
	return nil
}

func getJWKSKeys(ctx *fiber.Ctx) error {
	signingKey := state.StarState.SigningKey
	publicKey, err := signingKey.GetPublicKey()
	if err != nil {
		return err
	}

	jwk := jose.JSONWebKey{
		Key:       publicKey,
		Use:       "sig",
		Algorithm: string(crypto.JwtAlgorithm),
		KeyID:     crypto.JwtKeyID,
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	ctx.JSON(jwks)
	return nil
}

func getChangePassword(ctx *fiber.Ctx) error {
	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	if state.StarState.Config.Url != "" {
		hostname = state.StarState.Config.Url
	}

	return ctx.Redirect(hostname + "/#change-password", 303);
}