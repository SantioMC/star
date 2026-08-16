package oauth

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

func writeTokenResponse(ctx fiber.Ctx, response tokenResponse) error {
	ctx.Set(fiber.HeaderCacheControl, "no-store")
	ctx.Set("Pragma", "no-cache")

	return ctx.JSON(response)
}

func parseForm(ctx fiber.Ctx) (url.Values, bool, error) {
	contentType, _, _ := strings.Cut(ctx.Get(fiber.HeaderContentType), ";")
	if strings.TrimSpace(contentType) != fiber.MIMEApplicationForm {
		return nil, false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Requests must be sent as application/x-www-form-urlencoded")
	}

	form, err := url.ParseQuery(string(ctx.Body()))
	if err != nil {
		return nil, false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Malformed request body")
	}

	for key, values := range form {
		if len(values) > 1 {
			return nil, false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Request parameters must not be included more than once: "+key)
		}
	}

	return form, true, nil
}

func Token(ctx fiber.Ctx) error {
	form, ok, err := parseForm(ctx)
	if !ok {
		return err
	}

	application, ok, err := authenticateClient(ctx, form)
	if !ok {
		return err
	}

	switch form.Get("grant_type") {
	case "authorization_code":
		return authorizationCodeGrant(ctx, form, application)
	case "refresh_token":
		return refreshTokenGrant(ctx, form, application)
	case "":
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing grant_type")
	default:
		return writeTokenError(ctx, 400, Oauth2ErrorUnsupportedGrantType, "Only the authorization_code and refresh_token grant types are supported")
	}
}
