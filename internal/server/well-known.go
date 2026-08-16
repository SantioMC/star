package server

import (
	"fmt"

	"github.com/go-jose/go-jose/v4"
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/oauth"
)

func wellKnownRouter(router fiber.Router) {
	router.Get("/jwks.json", getJwks)
	router.Get("/openid-configuration", getOpenIDConfiguration)
}

func getJwks(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	keys := star.Database.GetAllSigningKeys(ctx)

	var jwks []jose.JSONWebKey
	for _, key := range keys {
		jwk, err := key.GetJWK()
		if err != nil {
			return err
		}

		jwks = append(jwks, jwk)
	}

	return ctx.JSON(jose.JSONWebKeySet{
		Keys: jwks,
	})
}

func getOpenIDConfiguration(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	return ctx.JSON(openid.DiscoveryDocument{
		Issuer:                            star.Config.ServerURL,
		JwksURI:                           fmt.Sprintf("%s/.well-known/jwks.json", star.Config.ServerURL),
		AuthorizationEndpoint:             fmt.Sprintf("%s/oauth/authorize", star.Config.ServerURL),
		TokenEndpoint:                     fmt.Sprintf("%s/oauth/token", star.Config.ServerURL),
		RevocationEndpoint:                fmt.Sprintf("%s/oauth/revoke", star.Config.ServerURL),
		UserinfoEndpoint:                  fmt.Sprintf("%s/oidc/userinfo", star.Config.ServerURL),
		DeviceAuthorizationEndpoint:       "",
		RegistrationEndpoint:              "",
		ScopesSupported:                   openid.SupportedScopes,
		ResponseTypesSupported:            oauth.SupportedResponseTypes,
		GrantTypesSupported:               oauth.SupportedGrantTypes,
		CodeChallengeMethodsSupported:     []string{oauth.CodeChallengeMethodS256},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post", "none"},
		SubjectTypesSupported:             []string{"public"},
		ClaimsSupported:                   []string{"aud", "auth_time", "azp", "email", "email_verified", "exp", "family_name", "given_name", "iat", "iss", "name", "nickname", "nonce", "picture", "preferred_username", "sid", "sub", openid.RolesClaim},
		MfaChallengeEndpoint:              "",
		ResponseModesSupported:            oauth.SupportedResponseModes,
		RequestParameterSupported:         false,
		IDTokenSigningAlgValuesSupported: []string{
			string(jose.ES512),
		},
	})
}
