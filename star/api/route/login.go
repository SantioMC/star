package route

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/state"
)

func LoginRouter(r fiber.Router) {
	r.Get("/:provider", limiter.New(helper.HeavyLimiter), loginCallback)
}

func loginCallback(ctx *fiber.Ctx) error {
	provider, err := state.StarState.Database.SocialStore.GetProviderByName(ctx.Params("provider", ""))
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

	app, err := state.StarState.Database.SocialStore.UseState(userState)
	if err != nil {
		slog.Error("user attempted to login with invalid state", "state", userState)
		return helper.NewHttpError(400, "unknown state provided")
	}

	hostname := string(ctx.Request().URI().Scheme()) + "://" + ctx.Hostname()
	token, err := provider.ExchangeCode(hostname, code)
	if err != nil {
		return err
	}

	_, _, err = state.StarState.Database.UserStore.CreateUserFromSocial(provider, token.AccessToken)
	if err != nil {
		return err
	}

	// todo: redirect
	return ctx.Redirect(app.Url)
}