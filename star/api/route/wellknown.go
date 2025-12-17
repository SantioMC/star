package route

import (
	"strings"

	"github.com/go-jose/go-jose/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/state"
)

func WellKnownRouter(r fiber.Router) {
	r.Get("security.txt", getWellKnownSecurity)
	r.Get("jwks.json", getJWKSKeys)
}

func getWellKnownSecurity(ctx *fiber.Ctx) error {
	ctx.SendString(strings.TrimSpace(
	"Contact: mailto:business@santio.me\n" +
	"Expires: 2027-1-01T00:00:00.000Z\n" +
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