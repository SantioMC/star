package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func AppRouter(r fiber.Router) {
	r.Get("/", getAllApps)
	r.Get("/:app", getApp)
	r.Get("/:app/logins", getAppLogins)
}

func getAllApps(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadApps); err != nil {
		return err
	}

	apps, err := service.GetAllApps()
	if err != nil {
		return err
	}

	return ctx.JSON(apps)
}

func getApp(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	return ctx.JSON(app)
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
