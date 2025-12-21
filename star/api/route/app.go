package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

func AppRouter(r fiber.Router) {
	r.Get("/:app/logins", getAppLogins)
}

func getAppLogins(ctx *fiber.Ctx) error {
	database := state.StarState.Database

	_, err := getAppFromCtx(ctx)
	if err != nil {
		return err
	}

	// todo: restrict which providers can login to which apps
	providers, err := database.SocialStore.GetProviders()
	var logins []string

	for _, provider := range providers {
		logins = append(logins, provider.Name)
	}

	ctx.JSON(AppLoginsResponse{
		Internal: true,
		Providers: logins,
	})

	return nil
}

func getAppFromCtx(ctx *fiber.Ctx) (*model.Application, error) {
	return state.StarState.Database.AppStore.GetAppBySlug(ctx.Params("app", ""))
}

type AppLoginsResponse struct {
	Internal  bool     `json:"user_pass_login"`
	Providers []string `json:"providers"`
}
