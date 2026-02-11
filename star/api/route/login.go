package route

import (
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/api/middleware"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state"
)

var loginClaims []oidc.Claim = []oidc.Claim{
	oidc.ClaimEmail, oidc.ClaimRoles, oidc.ClaimProfile,
}

func LoginRouter(r fiber.Router) {
	limiter := limiter.New(helper.HeavyLimiter)

	r.Post("/app/:app/login", limiter, middleware.RequireCaptcha, login)
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

	loginState, err := service.UseState(userState)
	if err != nil {
		slog.Error("user attempted to login with invalid state", "state", userState, "err", err)
		return helper.NewHttpError(400, "unknown state provided, please retry")
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	if state.StarState.Config.Url != "" {
		hostname = state.StarState.Config.Url
	}

	token, err := provider.ExchangeCode(hostname, code)
	if err != nil {
		return err
	}

	user, _, err := service.CreateUserFromSocial(provider, token.AccessToken)
	if err != nil {
		return err
	}

	if err := postLoginChecks(user); err != nil {
		return err
	}

	refreshToken, accessToken, _, err := service.CreateSession(
		user,
		loginState.Application,
		nil,
		loginState.UserAgent,
		loginState.Address,
		loginClaims,
	)

	if err != nil {
		return err
	}

	if err := service.UpdateLastLogin(user.ID, time.Now()); err != nil {
		return err
	}

	callbackUri := loginState.Application.Callback
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

	if err := postLoginChecks(user); err != nil {
		return err
	}

	userAgent := ""
	if body.UserAgent != nil && strings.TrimSpace(*body.UserAgent) != "" && len(*body.UserAgent) < 256 {
		userAgent = *body.UserAgent
	}

	session, accessToken, expiresAt, err := service.CreateSession(
		user,
		app,
		body.Fingerprint,
		userAgent,
		ctx.IP(),
		loginClaims,
	)

	if err != nil {
		return err
	}

	if err := service.UpdateLastLogin(user.ID, time.Now()); err != nil {
		return err
	}

	ctx.JSON(LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  session.Token,
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

	userAgent := ctx.Get(fiber.HeaderUserAgent)
	address := ctx.IP()

	code, err := service.CreateState(app, userAgent, address)
	if err != nil {
		return err
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	if state.StarState.Config.Url != "" {
		hostname = state.StarState.Config.Url
	}

	redirect := provider.GetCallbackURL(hostname, provider.Name)
	loginUrl, err := provider.GetLoginURL(code, redirect)
	if err != nil {
		return err
	}

	return ctx.Redirect(loginUrl)
}

func postLoginChecks(user *model.User) error {
	suspended, err := service.GetActiveSuspension(user.ID)
	if err != nil {
		return err
	}

	if suspended != nil {
		return helper.NewHttpError(403, "Your account has been suspended by an administrator")
	}

	if !user.EmailVerified {
		return helper.NewHttpError(403, "Your account has not had it's email verified")
	}

	return nil
}

type LoginRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	Fingerprint *string `json:"fingerprint"`
	UserAgent   *string `json:"userAgent"`
}

type LoginResponse struct {
	RefreshToken  string `json:"refresh_token,omitempty"`
	AccessToken   string `json:"access_token,omitempty"`
	ExpiresAt     int64  `json:"expires_at"`
	ApplicationID string `json:"app_id"`
	UserID        string `json:"user_id"`
}
