package oauth

import (
	"net/url"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/request"
)

var SupportedResponseTypes = []string{
	"code",
	"token",
	"code token",
	"code token id_token",
	"code id_token",
	"token id_token",
	"id_token",
}

var SupportedResponseModes = []string{
	"query",
	"form_post",
}

func authorizationError(ctx fiber.Ctx, err string) error {
	return ctx.Redirect().To("/oauth/authorize?error=" + url.QueryEscape(err))
}

func Authorize(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	if ctx.Query("error") != "" {
		return request.ShowFrontend(ctx)
	}

	// Get and validate params
	state := ctx.Query("state")
	responseType := ctx.Query("response_type")

	if !slices.Contains(SupportedResponseTypes, responseType) {
		return ctx.JSON(NewError(Oauth2ErrorUnsupportedResponseType, "Missing, unsupported or unknown response type", state))
	}

	clientId := ctx.Query("client_id")
	redirectUri := ctx.Query("redirect_uri")
	scopes := strings.Split(ctx.Query("scope"), " ")

	if clientId == "" {
		return ctx.JSON(NewError(Oauth2ErrorInvalidRequest, "Missing client_id", state))
	}

	if redirectUri == "" {
		return ctx.JSON(NewError(Oauth2ErrorInvalidRequest, "Missing redirect_uri, the authorization server requires one", state))
	}

	if len(scopes) == 0 || (len(scopes) == 1 && scopes[0] == "") {
		scopes = []string{"openid"}
	}

	// Get the application the user is trying to authorize with
	application := star.Database.GetApplicationByClientID(ctx, clientId)
	if application == nil {
		return authorizationError(ctx, "Invalid or unknown application, a client_id was specified but no matching application was found")
	}

	// Check if the application is configured to allow this redirect_uri
	if !slices.Contains(application.RedirectURIs, redirectUri) {
		return authorizationError(ctx, "The redirect_uri specified is not allowed for this application")
	}

	// PKCE flow
	if application.Kind == database.ApplicationKindPKCE && ctx.Query("code_challenge") == "" {
		return authorizationError(ctx, "PKCE flow requires a code_challenge parameter")
	}

	return request.ShowFrontend(ctx)
}
