package route

import (
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/verifier"
)

func ProviderRouter(router fiber.Router) {
	router.Post("/", createProvider)
	router.Delete("/:provider", deleteProvider)
	router.Patch("/:provider", patchProvider)

	router.Post("/:provider/rotate", rotateProvider)

	router.Post("/:provider/redirect", createRedirect)
	router.Delete("/:provider/redirect/:id", deleteRedirect)
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

	switch body.Kind {
	case "oauth":
		provider, err := service.CreateAppOAuthProvider(app)
		if err != nil {
			return err
		}

		return ctx.JSON(provider)
	default:
		return helper.NewHttpError(400, "unknown provider kind provided")
	}
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

	return ctx.JSON(fiber.Map{"success": true})
}

func patchProvider(ctx *fiber.Ctx) error {
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

	var body PatchProviderRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.ApplicationType != nil {
		appType := *body.ApplicationType

		if !slices.Contains(model.KnownAppTypes, appType) {
			return helper.NewHttpError(400, "invalid application type provided")
		}

		provider.ApplicationType = appType
	}

	if body.Fingerprinting != nil {
		provider.Fingerprinting = *body.Fingerprinting
	}

	if err := service.SaveProvider(provider); err != nil {
		return err
	}

	return ctx.JSON(provider)
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

func createRedirect(ctx *fiber.Ctx) error {
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

	var body CreateRedirectRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if err := verifier.VerifyURL(body.RedirectURI); err != nil {
		return helper.NewHttpError(400, err.Error())
	}

	if err := service.CreateRedirrect(provider, body.RedirectURI); err != nil {
		return err
	}

	return ctx.JSON(provider)
}

func deleteRedirect(ctx *fiber.Ctx) error {
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

	index, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return helper.NewHttpError(400, "invalid redirect id provided")
	}

	redirects := provider.RedirectURIs.Data()
	if len(redirects) <= index || index < 0 {
		return helper.NewHttpError(400, "redirect id out of range")
	}

	if err := service.DeleteRedirect(provider, redirects[index]); err != nil {
		return err
	}

	return ctx.JSON(provider)
}

func getProviderFromCtx(ctx *fiber.Ctx) (model.Application, model.OAuthProvider, error) {
	app, err := service.GetAppByID(ctx.Params("app", ""))
	if err != nil {
		return model.Application{}, model.OAuthProvider{}, err
	}

	if app == nil {
		return model.Application{}, model.OAuthProvider{}, helper.NewHttpError(404, "app not found")
	}

	provider, err := service.GetAppProviderByID(ctx.Params("provider", ""))
	if err != nil {
		return *app, model.OAuthProvider{}, err
	}

	if provider == nil || provider.ApplicationID != app.ID {
		return *app, model.OAuthProvider{}, helper.NewHttpError(404, "app not found")
	}

	return *app, *provider, nil
}

type CreateProviderRequest struct {
	Kind string `json:"kind"`
}

type PatchProviderRequest struct {
	ApplicationType *model.AppType `json:"application_type"`
	Fingerprinting  *bool          `json:"fingerprinting"`
}

type CreateRedirectRequest struct {
	RedirectURI string `json:"redirect_uri"`
}
