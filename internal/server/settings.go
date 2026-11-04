package server

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

const (
	minimumEmailCodeLength = 6
	maximumEmailCodeLength = 12
	minimumPasswordLength  = 8
	maximumPasswordLength  = 512
)

func settingsRouter(router fiber.Router) {
	readSettings := middleware.RequireScope(openid.ScopeAdminSettingsRead)
	writeSettings := middleware.RequireScope(openid.ScopeAdminSettingsWrite)

	router.Get("/", middleware.RequireAuth, readSettings, middleware.RequirePermission(permission.PermissionAdminReadSettings), getSettings)
	router.Patch("/", middleware.RequireAuth, writeSettings, middleware.RequirePermission(permission.PermissionAdminWriteSettings), updateSettings)
	router.Get("/disposable-domains", middleware.RequireAuth, readSettings, middleware.RequirePermission(permission.PermissionAdminReadSettings), getDisposableDomains)
}

func currentSettings(star *state.StarServer, ctx fiber.Ctx) (*database.Settings, error) {
	settings, err := star.Database.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		return nil, fiber.NewError(500, "This server has no settings configured")
	}

	return settings, nil
}

func getSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformSettings(*settings, *settings.GetPasswordSettings()))
}

func getDisposableDomains(ctx fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"domains": email.DisposableDomains(),
	})
}

func validateSettings(settings *database.Settings, passwordSettings *password.PasswordSetttings) error {
	if settings.EmailCodeLength < minimumEmailCodeLength || settings.EmailCodeLength > maximumEmailCodeLength {
		return fiber.NewError(400, "Email codes must be between 6 and 12 characters long")
	}

	if settings.EmailDomainMode == email.DomainModeWhitelist && len(settings.EmailDomains.Data) == 0 {
		return fiber.NewError(400, "A whitelist needs at least one domain, otherwise nobody could add an email address")
	}

	if passwordSettings.MinLength < minimumPasswordLength || passwordSettings.MinLength > maximumPasswordLength {
		return fiber.NewError(400, "The minimum password length must be between 8 and 512 characters")
	}

	if passwordSettings.MaxLength > maximumPasswordLength {
		return fiber.NewError(400, "The maximum password length may not exceed 512 characters")
	}

	if passwordSettings.MaxLength < passwordSettings.MinLength {
		return fiber.NewError(400, "The maximum password length must be greater than the minimum password length")
	}

	if passwordSettings.MinSymbols < 0 || passwordSettings.MinUpper < 0 || passwordSettings.MinLower < 0 {
		return fiber.NewError(400, "The required character counts may not be negative")
	}

	required := passwordSettings.MinSymbols + passwordSettings.MinUpper + passwordSettings.MinLower
	if required > passwordSettings.MinLength {
		return fiber.NewError(400, "The required characters add up to more than the minimum password length, no password could satisfy this policy")
	}

	return nil
}

func updateSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.UpdateSettingsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	before := *settings

	if request.AllowRegistration != nil {
		settings.AllowRegistration = *request.AllowRegistration
	}

	if request.AllowLogins != nil {
		settings.AllowLogins = *request.AllowLogins
	}

	if request.EmailCodeLength != nil {
		settings.EmailCodeLength = *request.EmailCodeLength
	}

	if request.EmailVerification != nil {
		settings.EmailVerification = *request.EmailVerification
	}

	if request.EmailDomainMode != nil {
		settings.EmailDomainMode = email.DomainMode(*request.EmailDomainMode)
	}

	if request.EmailDomains != nil {
		domains, err := email.NormalizeDomains(request.EmailDomains)
		if err != nil {
			return fiber.NewError(400, err.Error())
		}

		settings.EmailDomains.Data = domains
	}

	if request.BlockDisposableEmails != nil {
		settings.BlockDisposableEmails = *request.BlockDisposableEmails
	}

	currentPasswordSettings := *settings.GetPasswordSettings()
	passwordSettings := currentPasswordSettings

	if update := request.PasswordSettings; update != nil {
		if update.MinLength != nil {
			passwordSettings.MinLength = *update.MinLength
		}

		if update.MaxLength != nil {
			passwordSettings.MaxLength = *update.MaxLength
		}

		if update.MinSymbols != nil {
			passwordSettings.MinSymbols = *update.MinSymbols
		}

		if update.MinUpper != nil {
			passwordSettings.MinUpper = *update.MinUpper
		}

		if update.MinLower != nil {
			passwordSettings.MinLower = *update.MinLower
		}
	}

	if err := validateSettings(settings, &passwordSettings); err != nil {
		return err
	}

	if passwordSettings != currentPasswordSettings {
		encoded, err := json.Marshal(passwordSettings)
		if err != nil {
			return err
		}

		settings.PasswordSettings = string(encoded)
	}

	if err := star.Database.UpdateSettings(ctx, settings); err != nil {
		return err
	}

	cache.UpdateSettings(ctx, star)
	audit.Log(ctx, audit.SettingsUpdate, audit.Changes(before, *settings))

	return ctx.JSON(transformer.TransformSettings(*settings, passwordSettings))
}
