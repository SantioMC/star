package oauth

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/tubblelive/star/frontend"
	"github.com/tubblelive/star/star/service"
)

func sendError(ctx *fiber.Ctx, reason string) error {
	return ctx.Redirect("/oauth/authorize?error=" + url.QueryEscape(reason))
}

func HandleAuthorize(ctx *fiber.Ctx) error {
	if ctx.Query("error") != "" {
		return filesystem.SendFile(ctx, http.FS(frontend.FrontendFS), "build/index.html")
	}

	responseType := ctx.Query("response_type")
	if responseType != "code" {
		return sendError(ctx, "Invalid response type provided, expected 'code'")
	}

	clientId := ctx.Query("client_id")
	application, err := service.GetAppProviderByClientID(clientId)
	if err != nil {
		return err
	}

	if application == nil {
		return sendError(ctx, "Invalid client id provided")
	}

	callbackUri := ctx.Query("redirect_uri")
	allowedRedirects := application.RedirectURIs.Data()
	if callbackUri != "" && !slices.Contains(allowedRedirects, callbackUri) {
		return sendError(ctx, "Invalid redirect uri provided")
	}

	return filesystem.SendFile(ctx, http.FS(frontend.FrontendFS), "build/index.html")
}
