package route

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/verifier"
)

func AppRouter(r fiber.Router) {
	r.Get("/", getAllApps)
	r.Post("/", createApp)

	r.Get("/:app", getApp)
	r.Patch("/:app", updateApp)
	r.Delete("/:app", deleteApp)

	r.Get("/:app/logins", getAppLogins)
	r.Post("/:app/provider", createProvider)

	r.Delete("/:app/provider/:provider", deleteProvider)
	r.Post("/:app/provider/:provider/rotate", rotateProvider)
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

func createApp(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminCreateApps); err != nil {
		return err
	}

	var body CreateAppRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Slug != nil && *body.Slug != "" {
		if err := verifier.VerifyApplicationSlug(*body.Slug); err != nil {
			return helper.NewHttpError(400, err.Error())
		}
	}

	if err := verifier.VerifyURL(body.URL); err != nil {
		return helper.NewHttpError(400, err.Error())
	}

	existing, err := service.GetAppBySlug(*body.Slug)
	if err != nil {
		return err
	}

	if existing != nil {
		return helper.NewHttpError(400, "an application with the same slug already exists!")
	}

	app, err := service.CreateApplication(body.Name, body.Slug, body.URL)
	if err != nil {
		return err
	}

	return ctx.JSON(app)
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

func updateApp(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	if app.Internal {
		return helper.NewHttpError(403, "app is internal and can not be modified")
	}

	var body PatchAppRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Name != nil {
		app.Name = *body.Name
	}

	if body.Slug != nil && *body.Slug != "" {
		if err := verifier.VerifyApplicationSlug(*body.Slug); err != nil {
			return helper.NewHttpError(400, err.Error())
		}

		app.Slug = *body.Slug
	}

	if body.URL != nil {
		if err := verifier.VerifyURL(*body.URL); err != nil {
			return err
		}

		app.Url = *body.URL
	}

	if body.CallbackURL != nil {
		if err := verifier.VerifyURL(*body.CallbackURL); err != nil {
			return err
		}

		app.Callback = *body.CallbackURL
	}

	if err := service.UpdateApp(app); err != nil {
		return err
	}

	return ctx.JSON(app)
}

func deleteApp(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminDeleteApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app"))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	if app.Internal {
		return helper.NewHttpError(403, "app is internal and can not be deleted")
	}

	if err := service.DeleteApp(app.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
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

	return ctx.JSON(AppLoginsResponse{
		Internal:  true,
		Providers: logins,
	})
}

func createProvider(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, err := service.GetAppByID(ctx.Params("app", ""))
	if err != nil {
		return err
	}

	if app == nil {
		return helper.NewHttpError(404, "app not found")
	}

	if app.Internal {
		return helper.NewHttpError(403, "app is internal and can not be modified")
	}

	var body CreateProviderRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if (!slices.Contains(model.AllProviderKinds, body.Kind)) {
		return helper.NewHttpError(400, "unknown provider kind provided")
	}

	provider, err := service.CreateAppProvider(app, body.Kind)
	if err != nil {
		return err
	}

	return ctx.JSON(provider)
}

func deleteProvider(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, provider, err := getProviderFromCtx(ctx)
	if err != nil {
		return err
	}

	if app.Internal {
		return helper.NewHttpError(403, "app is internal and can not be modified")
	}

	if err := service.DeleteAppProvider(provider.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{ "success": true })
}

func rotateProvider(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyApps); err != nil {
		return err
	}

	app, provider, err := getProviderFromCtx(ctx)
	if err != nil {
		return err
	}

	if app.Internal {
		return helper.NewHttpError(403, "app is internal and can not be modified")
	}

	if err := service.RotateAppProvider(provider); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"secret": provider.Secret})
}

func getProviderFromCtx(ctx *fiber.Ctx) (model.Application, model.Provider, error) {
	app, err := service.GetAppByID(ctx.Params("app", ""))
	if err != nil {
		return model.Application{}, model.Provider{}, err
	}

	if app == nil {
		return model.Application{}, model.Provider{}, helper.NewHttpError(404, "app not found")
	}

	provider, err := service.GetAppProviderByID(ctx.Params("provider", ""))
	if err != nil {
		return *app, model.Provider{}, err
	}

	if provider == nil || provider.ApplicationID != app.ID {
		return *app, model.Provider{}, helper.NewHttpError(404, "app not found")
	}

	return *app, *provider, nil
}

type AppLoginsResponse struct {
	Internal  bool     `json:"user_pass_login"`
	Providers []string `json:"providers"`
}

type CreateAppRequest struct {
	Name string  `json:"name"`
	Slug *string `json:"slug"`
	URL  string  `json:"url"`
}

type PatchAppRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	URL         *string `json:"url"`
	CallbackURL *string `json:"callback_url"`
}

type CreateProviderRequest struct {
	Kind model.ProviderKind `json:"kind"`
}

