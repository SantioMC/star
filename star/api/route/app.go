package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/service"
)

func AppRouter(r fiber.Router) {
	r.Get("/:app/logins", getAppLogins)
}

func getAppLogins(ctx *fiber.Ctx) error {
	_, err := service.GetAppBySlug(ctx.Params("app", ""))
	if err != nil {
		return err
	}

	// todo: restrict which providers can login to which apps
	providers, err := service.GetProviders()
	logins := []string{}

	for _, provider := range providers {
		logins = append(logins, provider.Name)
	}

	ctx.JSON(AppLoginsResponse{
		Internal: true,
		Providers: logins,
	})

	return nil
}

type AppLoginsResponse struct {
	Internal  bool     `json:"user_pass_login"`
	Providers []string `json:"providers"`
}
