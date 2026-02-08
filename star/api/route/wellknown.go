package route

import (
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/tubblelive/star"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/state"
)

func WellKnownRouter(r fiber.Router) {
	cache := cache.New(cache.Config{
		Expiration:   24 * time.Hour,
		CacheControl: true,
	})

	r.Get("security.txt", getWellKnownSecurity)
	r.Get("change-password", getChangePassword)
	r.Get("jwks.json", cache, getJWKSKeys)
	// r.Get("openid-configuration", cache, getOpenidConfig)
}

func getWellKnownSecurity(ctx *fiber.Ctx) error {
	expiresAt := star.StartedAt.Add(time.Hour * 24 * 365)

	email := state.StarState.Config.SecurityEmail
	if email == "" {
		// No contact email, 404 this page
		return nil
	}

	return ctx.SendString(strings.TrimSpace(
		"Contact: mailto:" + email + "\n" +
			"Expires: " + expiresAt.UTC().Format(time.RFC3339Nano) + "\n" +
			"Preferred-Languages: en",
	))
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

	return ctx.Redirect(hostname+"/#change-password", 303)
}

// func getOpenidConfig(ctx *fiber.Ctx) error {
// 	url := strings.TrimSuffix(state.StarState.Config.Url, "/")

// 	userInfo := url + "/oauth/userinfo"
// 	registerEndpoint := url + "/register"
// 	jwksEndpoint := url + "/.well-known/jwks.json"

// 	return ctx.JSON(oidc.OpenIDDiscovery{
// 		Issuer: url,

// 		AuthEndpoint: url + "/oauth/authorize",
// 		TokenEndpoint: url + "/oauth/token",

// 		RegisterEndpoint: &registerEndpoint,
// 		UserInfoEndpoint: &userInfo,
// 		JwksEndpoint: &jwksEndpoint,

// 		SupportedTokenAuth: &[]string{"client_secret_basic"},
// 		SupportedTokenAlg: &[]string{"ES512"},

// 		SupportedScopes: scopes.DefaultScopes,
// 	})
// }
