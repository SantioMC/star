package server

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func brandingRouter(router fiber.Router) {
	writeScope := middleware.RequireScope(openid.ScopeAdminSettingsWrite)
	writePermission := middleware.RequirePermission(permission.PermissionAdminWriteSettings)

	router.Get("/", getBranding)
	router.Get("/logo", getBrandingLogo)

	router.Post("/preview", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminSettingsRead), middleware.RequirePermission(permission.PermissionAdminReadSettings), createEmailPreview)
	router.Get("/preview", renderEmailPreview)

	router.Patch("/", middleware.RequireAuth, writeScope, writePermission, updateBranding)
	router.Put("/logo", middleware.RequireAuth, writeScope, writePermission, uploadBrandingLogo)
	router.Delete("/logo", middleware.RequireAuth, writeScope, writePermission, deleteBrandingLogo)
}

func currentBranding(star *state.StarServer, ctx fiber.Ctx) (*database.Branding, error) {
	current, err := star.Database.GetBranding(ctx)
	if err != nil {
		return nil, err
	}

	if current == nil {
		return nil, fiber.NewError(500, "This server has no branding configured")
	}

	return current, nil
}

func respondBranding(star *state.StarServer, ctx fiber.Ctx) error {
	cache.UpdateBranding(ctx, star)
	return getBranding(ctx)
}

func getBranding(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	current := star.Branding.Load()
	if current == nil {
		return fiber.NewError(500, "This server has no branding configured")
	}

	ctx.Set(fiber.HeaderCacheControl, "no-cache")
	return ctx.JSON(transformer.TransformBranding(current, star.Config.ServerURL))
}

func getBrandingLogo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	logo, err := star.Database.GetBrandingLogo(ctx)
	if err != nil {
		return err
	}

	if logo == nil {
		return fiber.ErrNotFound
	}

	etag := `"` + strconv.FormatInt(logo.UpdatedAt.UnixNano(), 36) + `"`
	ctx.Set(fiber.HeaderETag, etag)
	ctx.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	ctx.Set(fiber.HeaderContentSecurityPolicy, "default-src 'none'")

	if ctx.Get(fiber.HeaderIfNoneMatch) == etag {
		return ctx.SendStatus(fiber.StatusNotModified)
	}

	ctx.Set(fiber.HeaderContentType, logo.Type)
	return ctx.Send(logo.Data)
}

func cleared(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func applyBranding(current *database.Branding, request *types.UpdateBrandingRequest) error {
	for _, font := range []*string{request.HeadingFont, request.BodyFont} {
		if font != nil && !branding.ValidFont(*font) {
			return fiber.NewError(400, "Fonts must be the name of a Google Font, using only letters, numbers and spaces")
		}
	}

	if request.Name != nil {
		current.Name = *request.Name
	}

	if request.AccentColor != nil {
		current.AccentColor = *request.AccentColor
	}

	if request.HeadingFont != nil {
		current.HeadingFont = *request.HeadingFont
	}

	if request.BodyFont != nil {
		current.BodyFont = *request.BodyFont
	}

	if request.Layout != nil {
		current.Layout = *request.Layout
	}

	if request.SupportEmail != nil {
		current.SupportEmail = cleared(*request.SupportEmail)
	}

	if request.PrivacyURL != nil {
		current.PrivacyURL = cleared(*request.PrivacyURL)
	}

	if request.TermsURL != nil {
		current.TermsURL = cleared(*request.TermsURL)
	}

	return nil
}

func updateBranding(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.UpdateBrandingRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	current, err := currentBranding(star, ctx)
	if err != nil {
		return err
	}

	if err := applyBranding(current, &request); err != nil {
		return err
	}

	if err := star.Database.UpdateBranding(ctx, current); err != nil {
		return err
	}

	return respondBranding(star, ctx)
}

func uploadBrandingLogo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	data := ctx.Body()

	if len(data) == 0 {
		return fiber.NewError(400, "No logo was provided")
	}

	if len(data) > 512*1024 {
		return fiber.NewError(413, "Logos may not be larger than 512 KB")
	}

	contentType, allowed := branding.DetectLogoType(data)
	if !allowed {
		return fiber.NewError(415, "Logos must be a PNG, JPEG or WebP image")
	}

	if err := star.Database.UpdateBrandingLogo(ctx, data, contentType); err != nil {
		return err
	}

	return respondBranding(star, ctx)
}

func deleteBrandingLogo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	if err := star.Database.DeleteBrandingLogo(ctx); err != nil {
		return err
	}

	return respondBranding(star, ctx)
}
