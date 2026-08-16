package oauth

import "github.com/gofiber/fiber/v3"

type Oauth2Error string

const (
	Oauth2ErrorInvalidRequest          Oauth2Error = "invalid_request"
	Oauth2ErrorInvalidClient           Oauth2Error = "invalid_client"
	Oauth2ErrorInvalidGrant            Oauth2Error = "invalid_grant"
	Oauth2ErrorUnauthorizedClient      Oauth2Error = "unauthorized_client"
	Oauth2ErrorAccessDenied            Oauth2Error = "access_denied"
	Oauth2ErrorUnsupportedResponseType Oauth2Error = "unsupported_response_type"
	Oauth2ErrorUnsupportedGrantType    Oauth2Error = "unsupported_grant_type"
	Oauth2ErrorUnsupportedTokenType    Oauth2Error = "unsupported_token_type"
	Oauth2ErrorInvalidScope            Oauth2Error = "invalid_scope"
	Oauth2ErrorServerError             Oauth2Error = "server_error"
	Oauth2ErrorTemporarilyUnavailable  Oauth2Error = "temporarily_unavailable"
)

type errorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
	State       string `json:"state,omitempty"`
}

func NewError(kind Oauth2Error, message string, state string) errorResponse {
	return errorResponse{
		Error:       string(kind),
		Description: message,
		State:       state,
	}
}

func writeTokenError(ctx fiber.Ctx, status int, kind Oauth2Error, description string) error {
	ctx.Set(fiber.HeaderCacheControl, "no-store")
	ctx.Set("Pragma", "no-cache")

	return ctx.Status(status).JSON(NewError(kind, description, ""))
}

func invalidClient(ctx fiber.Ctx, usedBasic bool, description string) error {
	status := 400
	if usedBasic {
		status = 401
		ctx.Set(fiber.HeaderWWWAuthenticate, `Basic realm="star"`)
	}

	return writeTokenError(ctx, status, Oauth2ErrorInvalidClient, description)
}
