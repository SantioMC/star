package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

var loginClaims []oidc.Claim = []oidc.Claim{
	oidc.ClaimEmail, oidc.ClaimRoles,
}

func AppRouter(r fiber.Router) {
	r.Post("/:app/login", limiter.New(helper.HeavyLimiter), login)
	r.Get("/:app/login/:provider", limiter.New(helper.HeavyLimiter), loginWithProvider)
	r.Get("/callback/:provider", limiter.New(helper.HeavyLimiter), loginWithProvider)
}

func login(ctx *fiber.Ctx) error {
	body := new(LoginRequest)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	app, err := getAppFromCtx(ctx)
	if err != nil {
		return helper.NewHttpError(404, "Invalid application slug provided")
	}

	userStore := state.StarState.Database.UserStore
	user, err := userStore.ValidateLogin(body.Username, body.Password)
	if err != nil {
		return helper.NewHttpError(401, "Invalid username or password")
	}

	refreshToken, accessToken, expiresAt, err := userStore.CreateRefreshToken(user, app, loginClaims)
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
	_, err := getAppFromCtx(ctx)
	if err != nil {
		return helper.NewHttpError(404, "Invalid application slug provided")
	}

	provider, err := state.StarState.Database.SocialStore.GetProviderByName(ctx.Params("provider", ""))
	if err != nil {
		return helper.NewHttpError(404, "Invalid provider name provided")
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	state := crypto.CreateRandString(16) // todo: bind state to app so we can provision RT/AT appropriately
	redirect := provider.GetCallbackURL(hostname, provider.Name)

	return ctx.Redirect(provider.GetLoginURL(state, redirect))
}

func getAppFromCtx(ctx *fiber.Ctx) (*model.Application, error) {
	return state.StarState.Database.AppStore.GetAppBySlug(ctx.Params("app", ""))
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
