package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
)

type providerPreset struct {
	Name             string   `json:"name"`
	Icon             string   `json:"icon"`
	Scopes           []string `json:"scopes,omitempty"`
	URL              *string  `json:"oidc_url,omitempty"`
	AuthorizationURL *string  `json:"authorization_url,omitempty"`
	TokenURL         *string  `json:"token_url,omitempty"`
	UserinfoURL      *string  `json:"userinfo_url,omitempty"`
	TrustEmail       bool     `json:"trust_email"`
}

var providerPresets = []providerPreset{
	{
		Name:       "Google",
		Icon:       "brand-google",
		URL:        new("https://accounts.google.com/.well-known/openid-configuration"),
		TrustEmail: true,
	},
	{
		Name:             "GitHub",
		Icon:             "brand-github",
		Scopes:           []string{"read:user", "user:email"},
		AuthorizationURL: new("https://github.com/login/oauth/authorize"),
		TokenURL:         new("https://github.com/login/oauth/access_token"),
		UserinfoURL:      new("https://api.github.com/user"),
	},
	{
		Name:       "GitLab",
		Icon:       "brand-gitlab",
		URL:        new("https://gitlab.com/.well-known/openid-configuration"),
		TrustEmail: true,
	},
	{
		Name:             "Discord",
		Icon:             "brand-discord",
		Scopes:           []string{"identify", "email"},
		AuthorizationURL: new("https://discord.com/oauth2/authorize"),
		TokenURL:         new("https://discord.com/api/oauth2/token"),
		UserinfoURL:      new("https://discord.com/api/users/@me"),
	},
	{
		Name: "Microsoft",
		Icon: "brand-windows",
		URL:  new("https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"),
	},
	{
		Name:       "Apple",
		Icon:       "brand-apple",
		Scopes:     []string{"name", "email"},
		URL:        new("https://appleid.apple.com/.well-known/openid-configuration"),
		TrustEmail: true,
	},
	{
		Name:             "Facebook",
		Icon:             "brand-facebook",
		Scopes:           []string{"email", "public_profile"},
		AuthorizationURL: new("https://www.facebook.com/v19.0/dialog/oauth"),
		TokenURL:         new("https://graph.facebook.com/v19.0/oauth/access_token"),
		UserinfoURL:      new("https://graph.facebook.com/me"),
	},
	{
		Name: "Twitch",
		Icon: "brand-twitch",
		URL:  new("https://id.twitch.tv/oauth2/.well-known/openid-configuration"),
	},
}

var defaultProviderScopes = []string{"openid", "profile", "email"}

func providerRouter(router fiber.Router) {
	readProviders := middleware.RequireScope(openid.ScopeAdminProvidersRead)
	writeProviders := middleware.RequireScope(openid.ScopeAdminProvidersWrite)

	router.Get("/", middleware.RequireAuth, readProviders, middleware.RequirePermission(permission.PermissionAdminReadProviders), getProviders)
	router.Post("/", middleware.RequireAuth, writeProviders, middleware.RequirePermission(permission.PermissionAdminCreateProviders), createProvider)

	router.Get("/presets", middleware.RequireAuth, readProviders, middleware.RequirePermission(permission.PermissionAdminReadProviders), getProviderPresets)

	router.Get("/:provider", middleware.RequireAuth, readProviders, middleware.RequirePermission(permission.PermissionAdminReadProviders), getProvider)
	router.Patch("/:provider", middleware.RequireAuth, writeProviders, middleware.RequirePermission(permission.PermissionAdminWriteProviders), updateProvider)
	router.Delete("/:provider", middleware.RequireAuth, writeProviders, middleware.RequirePermission(permission.PermissionAdminDeleteProviders), deleteProvider)
}

func getProviders(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	return ctx.JSON(transformer.TransformProviders(star.Database.GetProviders(ctx)))
}

func getProviderPresets(ctx fiber.Ctx) error {
	return ctx.JSON(providerPresets)
}

func getProvider(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	provider := star.Database.GetProviderByID(ctx, ctx.Params("provider"))
	if provider == nil {
		return fiber.NewError(404, "Provider not found")
	}

	return ctx.JSON(transformer.TransformProvider(*provider))
}

func validateProviderEndpoints(provider *database.Provider) error {
	if provider.OIDCUrl != nil {
		return nil
	}

	if provider.AuthorizationURL == nil || provider.TokenURL == nil || provider.UserinfoURL == nil {
		return fiber.NewError(400, "Providers require either an oidc_url or all of authorization_url, token_url and userinfo_url")
	}

	return nil
}

func createProvider(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.CreateProviderRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if star.Database.GetProviderByName(ctx, request.Name) != nil {
		return fiber.NewError(400, "A provider with this name already exists")
	}

	icon := "login"
	if request.Icon != nil {
		icon = *request.Icon
	}

	order := 0
	if request.Order != nil {
		order = *request.Order
	}

	scopes := defaultProviderScopes
	if len(request.Scopes) > 0 {
		scopes = request.Scopes
	}

	provider := &database.Provider{
		ID:               database.CreateID(database.ProviderKey),
		Name:             request.Name,
		Icon:             icon,
		Order:            order,
		ClientID:         request.ClientID,
		ClientSecret:     request.ClientSecret,
		Scopes:           pq.StringArray(scopes),
		OIDCUrl:          request.OIDCUrl,
		AuthorizationURL: request.AuthorizationURL,
		TokenURL:         request.TokenURL,
		UserinfoURL:      request.UserinfoURL,
		TrustEmail:       request.TrustEmail,
		CreatedAt:        time.Now(),
	}

	if err := validateProviderEndpoints(provider); err != nil {
		return err
	}

	if err := star.Database.CreateProvider(ctx, provider); err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformProvider(*provider))
}

func updateProvider(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.UpdateProviderRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	provider := star.Database.GetProviderByID(ctx, ctx.Params("provider"))
	if provider == nil {
		return fiber.NewError(404, "Provider not found")
	}

	if request.Name != nil && *request.Name != provider.Name {
		if star.Database.GetProviderByName(ctx, *request.Name) != nil {
			return fiber.NewError(400, "A provider with this name already exists")
		}

		provider.Name = *request.Name
	}

	if request.Icon != nil {
		provider.Icon = *request.Icon
	}

	if request.Order != nil {
		provider.Order = *request.Order
	}

	if request.ClientID != nil {
		provider.ClientID = *request.ClientID
	}

	if request.ClientSecret != nil {
		provider.ClientSecret = *request.ClientSecret
	}

	if request.Scopes != nil {
		provider.Scopes = pq.StringArray(*request.Scopes)
	}

	if request.OIDCUrl != nil {
		provider.OIDCUrl = request.OIDCUrl
	}

	if request.AuthorizationURL != nil {
		provider.AuthorizationURL = request.AuthorizationURL
	}

	if request.TokenURL != nil {
		provider.TokenURL = request.TokenURL
	}

	if request.UserinfoURL != nil {
		provider.UserinfoURL = request.UserinfoURL
	}

	if request.TrustEmail != nil {
		provider.TrustEmail = *request.TrustEmail
	}

	if err := validateProviderEndpoints(provider); err != nil {
		return err
	}

	if err := star.Database.UpdateProvider(ctx, provider); err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformProvider(*provider))
}

func deleteProvider(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	provider := star.Database.GetProviderByID(ctx, ctx.Params("provider"))
	if provider == nil {
		return fiber.NewError(404, "Provider not found")
	}

	if err := star.Database.DeleteProviderByID(ctx, provider.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
