package oauth

import (
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/request"
)

var SupportedResponseTypes = []string{
	"code",
}

var SupportedResponseModes = []string{
	"query",
}

var SupportedGrantTypes = []string{
	"authorization_code",
	"refresh_token",
	DeviceCodeGrantType,
}

func authorizationError(ctx fiber.Ctx, err string) error {
	return ctx.Redirect().To("/oauth/authorize?error=" + url.QueryEscape(err))
}

func redirectError(ctx fiber.Ctx, redirectURI string, kind Oauth2Error, description string, state string) error {
	target, err := url.Parse(redirectURI)
	if err != nil {
		slog.Warn("invalid redirect_uri specified for app uri", "uri", redirectURI)
		return authorizationError(ctx, "The redirect_uri specified is not a valid URI")
	}

	query := target.Query()
	query.Set("error", string(kind))
	query.Set("error_description", description)

	if state != "" {
		query.Set("state", state)
	}

	target.RawQuery = query.Encode()
	return ctx.Redirect().To(target.String())
}

func Authorize(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	if ctx.Query("error") != "" {
		return request.ShowFrontend(ctx)
	}

	query, err := url.ParseQuery(string(ctx.Request().URI().QueryString()))
	if err != nil {
		return authorizationError(ctx, "Malformed query parameters")
	}

	for key, values := range query {
		if len(values) > 1 {
			return authorizationError(ctx, "Request parameters must not be included more than once: "+key)
		}
	}

	clientID := ctx.Query("client_id")
	if clientID == "" {
		return authorizationError(ctx, "Missing client_id")
	}

	application := star.Database.GetApplicationByClientID(ctx, clientID)
	if application == nil {
		return authorizationError(ctx, "Invalid or unknown application, a client_id was specified but no matching application was found")
	}

	redirectURI := ctx.Query("redirect_uri")
	if redirectURI == "" {
		return authorizationError(ctx, "Missing redirect_uri, the authorization server requires one")
	}

	if !AllowedRedirectURI(application, redirectURI) {
		slog.Warn("invalid redirect_uri specified for app uri", "uri", redirectURI)
		return authorizationError(ctx, "The redirect_uri specified is not allowed for this application")
	}

	state := ctx.Query("state")
	responseType := ctx.Query("response_type")

	if responseType == "" {
		return redirectError(ctx, redirectURI, Oauth2ErrorInvalidRequest, "Missing response_type", state)
	}

	if !slices.Contains(SupportedResponseTypes, responseType) {
		return redirectError(ctx, redirectURI, Oauth2ErrorUnsupportedResponseType, "Unsupported or unknown response type, only 'code' is supported", state)
	}

	scopes := strings.Fields(ctx.Query("scope"))
	if len(scopes) == 0 {
		scopes = []string{openid.ScopeOpenID}
	}

	resolution := openid.ResolveScopes(ctx, star.Database, application, scopes)
	if len(resolution.Granted) == 0 {
		return redirectError(ctx, redirectURI, Oauth2ErrorInvalidScope, "None of the requested scopes are supported", state)
	}

	challenge := ctx.Query("code_challenge")
	method := ctx.Query("code_challenge_method")

	if kind, description, ok := ValidateCodeChallenge(application, challenge, method); !ok {
		return redirectError(ctx, redirectURI, kind, description, state)
	}

	return request.ShowFrontend(ctx)
}
