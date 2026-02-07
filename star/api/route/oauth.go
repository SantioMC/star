package route

import (
	"net/url"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/api/route/oauth"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func OAuthRouter(r fiber.Router) {
	rateLimiter := limiter.New(helper.HeavyLimiter)

	// Nonstandard
	r.Get("application/:id", getOAuthApplication)
	r.Post("login/code", getOAuthAuthCode)

	// OAuth2 standard
	r.Get("authorize", oauth.HandleAuthorize)
	r.Post("token", rateLimiter, oauth.HandleOidcToken)
}

func getOAuthApplication(ctx *fiber.Ctx) error {
	unescaped, err := url.QueryUnescape(ctx.Params("id"))
	if err != nil {
		return err
	}

	provider, err := service.GetAppProviderByClientID(unescaped)
	if err != nil {
		return err
	}

	if provider == nil {
		return helper.NewHttpError(404, "provider not found")
	}

	application, err := service.GetAppByID(provider.ApplicationID)
	if err != nil {
		return err
	}

	if application == nil {
		return helper.NewHttpError(404, "application not found")
	}

	// todo: add explicit authorization
	return ctx.JSON(PublicApplicationInfo{
		ID:           application.ID,
		Name:         application.Name,
		ImplicitAuth: true,
	})
}

func getOAuthAuthCode(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeCreateAuthorization); err != nil {
		return err
	}

	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	user, err := service.GetUserByID(userId)
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(404, "user not found")
	}

	var body CreateAuthRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	//todo: add ability to restrict authorization to apps
	app, err := service.GetAppByID(body.ApplicationID)
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "application not found")
	}

	allowedRedirects := app.OauthProvider.RedirectURIs.Data()
	if len(allowedRedirects) == 0 {
		return helper.NewHttpError(400, "application has no redirect uris configured")
	}

	redirect := allowedRedirects[0]
	if body.RedirectURI != nil && *body.RedirectURI != "" {
		redirect = *body.RedirectURI
	}

	if !slices.Contains(allowedRedirects, redirect) {
		return helper.NewHttpError(400, "invalid redirect uri provided")
	}

	scopes := oidc.DefaultScopes
	if body.Scopes != nil {
		scopes = *body.Scopes

		for i, scope := range scopes {
			if !slices.Contains(oidc.DefaultScopes, scope) {
				scopes = append(scopes[:i], scopes[i+1:]...)
			}
		}

		scopes = slices.Compact(scopes)
		if len(scopes) == 0 {
			return helper.NewHttpError(400, "invalid scope provided")
		}
	}

	code, err := service.CreateAuthRequest(
		app,
		user,
		scopes,
		redirect,
		body.State,
	)

	if err != nil {
		return err
	}

	return ctx.JSON(AuthorizationCodeResponse{
		Code:     code.Code,
		State:    body.State,
		Redirect: redirect,
	})
}

type PublicApplicationInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ImplicitAuth bool   `json:"implicit"`
}

type AuthorizationCodeResponse struct {
	Code     string  `json:"code"`
	State    *string `json:"state"`
	Redirect string  `json:"redirect"`
}

type CreateAuthRequest struct {
	ApplicationID string    `json:"application_id"`
	RedirectURI   *string   `json:"redirect_uri"`
	State         *string   `json:"state"`
	Scopes        *[]string `json:"scopes"`
}
