package route

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
)

var loginClaims []oidc.Claim = []oidc.Claim{
	oidc.ClaimEmail, oidc.ClaimRoles, oidc.ClaimProfile,
}

func LoginRouter(r fiber.Router) {
	limiter := limiter.New(helper.HeavyLimiter)

	r.Post("/app/:app/login", limiter, login)
	r.Get("/app/:app/login/:provider", limiter, loginWithProvider)
	r.Get("/callback/:provider", limiter, loginCallback)
}

func loginCallback(ctx *fiber.Ctx) error {
	provider, err := service.GetProviderByName(ctx.Params("provider", ""))
	if err != nil {
		return helper.NewHttpError(404, "Invalid provider name provided")
	}

	if !provider.IsValid() {
		return helper.NewHttpError(404, "Invalid provider provided")
	}

	code := ctx.Query("code")
	if code == "" {
		return helper.NewHttpError(400, "invalid token provided")
	}

	userState := ctx.Query("state")
	if userState == "" {
		// We want to make sure the state does exist, just for added security & handling
		return helper.NewHttpError(400, "invalid state provided")
	}

	app, err := service.UseState(userState)
	if err != nil {
		slog.Error("user attempted to login with invalid state", "state", userState)
		return helper.NewHttpError(400, "unknown state provided")
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	token, err := provider.ExchangeCode(hostname, code)
	if err != nil {
		return err
	}

	user, _, err := service.CreateUserFromSocial(provider, token.AccessToken)
	if err != nil {
		return err
	}

	suspended, err := service.GetActiveSuspension(user.ID)
	if err != nil {
		return err
	}

	if suspended != nil {
		return helper.NewHttpError(403, "Your account has been suspended by an administrator")
	}

	refreshToken, accessToken, _, err := service.CreateSession(user, app, nil, "", loginClaims)
	if err != nil {
		return err
	}

	callbackUri := app.Callback
	if strings.HasPrefix(callbackUri, "/") {
		// Use the current hostname
		callbackUri = strings.TrimSuffix(hostname, "/") + callbackUri
	}

	callbackUri += "?access_token=" + url.QueryEscape(accessToken)
	callbackUri += "&refresh_token=" + url.QueryEscape(refreshToken.Token)

	return ctx.Redirect(callbackUri)
}

func login(ctx *fiber.Ctx) error {
	body := new(LoginRequest)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	app, err := service.GetAppBySlug(ctx.Params("app", ""))
	if err != nil {
		return helper.NewHttpError(404, "Invalid application slug provided")
	}

	user, err := service.ValidateLogin(body.Username, body.Password)
	if err != nil {
		return helper.NewHttpError(401, "Invalid username or password")
	}

	suspended, err := service.GetActiveSuspension(user.ID)
	if err != nil {
		return err
	}

	if suspended != nil {
		return helper.NewHttpError(403, "Your account has been suspended by an administrator")
	}

	refreshToken, accessToken, expiresAt, err := service.CreateSession(user, app, nil, "", loginClaims)
	if err != nil {
		return err
	}

	ctx.JSON(LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken.Token,
		ExpiresAt:     expiresAt.Unix(),
		ApplicationID: app.ID,
		UserID:        user.ID,
	})

	return nil
}

func loginWithProvider(ctx *fiber.Ctx) error {
	app, err := service.GetAppBySlug(ctx.Params("app", ""))
	if err != nil {
		return helper.NewHttpError(404, "Invalid application slug provided")
	}

	provider, err := service.GetProviderByName(ctx.Params("provider", ""))
	if err != nil {
		return helper.NewHttpError(404, "Invalid provider name provided")
	}

	if !provider.IsValid() {
		return helper.NewHttpError(404, "Invalid provider provided")
	}

	state, err := service.CreateState(app)
	if err != nil {
		return err
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	redirect := provider.GetCallbackURL(hostname, provider.Name)

	loginUrl, err := provider.GetLoginURL(state, redirect)
	if err != nil {
		return err
	}

	return ctx.Redirect(loginUrl)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	RefreshToken  string `json:"refresh_token"`
	AccessToken   string `json:"access_token"`
	ExpiresAt     int64  `json:"expires_at"`
	ApplicationID string `json:"app_id"`
	UserID        string `json:"user_id"`
}
