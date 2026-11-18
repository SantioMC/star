package server

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
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
	router.Post("/email/test", middleware.RequireAuth, writeSettings, middleware.RequirePermission(permission.PermissionAdminWriteSettings), middleware.RateLimit(5), testEmailSettings)
	router.Post("/captcha/test", middleware.RequireAuth, writeSettings, middleware.RequirePermission(permission.PermissionAdminWriteSettings), middleware.RateLimit(10), testCaptchaSettings)
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

	if settings.MinimumAge < 0 || settings.MinimumAge > 120 {
		return fiber.NewError(400, "The minimum age must be between 0 and 120")
	}

	if settings.MinimumAge > 0 && !settings.TrackBirthdays {
		return fiber.NewError(400, "A minimum age needs birthday tracking to be enabled")
	}

	if settings.LockoutThreshold < 3 || settings.LockoutThreshold > 100 {
		return fiber.NewError(400, "The lockout threshold must be between 3 and 100 failed attempts")
	}

	if settings.LockoutWindowMinutes < 1 || settings.LockoutWindowMinutes > 1440 {
		return fiber.NewError(400, "The lockout window must be between 1 minute and 24 hours")
	}

	if settings.LockoutDurationMinutes < 1 || settings.LockoutDurationMinutes > 1440 {
		return fiber.NewError(400, "The lockout duration must be between 1 minute and 24 hours")
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

	if err := settings.GetEmailConfig().Validate(); err != nil {
		return err
	}

	if err := settings.GetCaptchaConfig().Validate(); err != nil {
		return err
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

	if request.TrackBirthdays != nil {
		settings.TrackBirthdays = *request.TrackBirthdays
	}

	if request.MinimumAge != nil {
		settings.MinimumAge = *request.MinimumAge
	}

	if request.LockoutThreshold != nil {
		settings.LockoutThreshold = *request.LockoutThreshold
	}

	if request.LockoutWindowMinutes != nil {
		settings.LockoutWindowMinutes = *request.LockoutWindowMinutes
	}

	if request.LockoutDurationMinutes != nil {
		settings.LockoutDurationMinutes = *request.LockoutDurationMinutes
	}

	applyEmailSettings(settings, &request)
	applyCaptchaSettings(settings, &request)

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

	captchaConfig := settings.GetCaptchaConfig()
	if captchaConfig.IsEnabled() && captchaConfig != before.GetCaptchaConfig() {
		if err := captchaConfig.CheckSecret(); err != nil {
			return err
		}
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
	audit.Log(
		ctx,
		audit.SettingsUpdate,
		audit.Changes(before, *settings),
		audit.With("secrets_changed", changedSecrets(&request)),
	)

	return ctx.JSON(transformer.TransformSettings(*settings, passwordSettings))
}

func testEmailSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.UpdateSettingsRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	applyEmailSettings(settings, &data)

	config := settings.GetEmailConfig()
	if !config.IsEnabled() {
		return fiber.NewError(400, "Choose an email provider before sending a test email")
	}

	if err := config.Validate(); err != nil {
		return err
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	primary := star.Database.GetPrimaryEmail(ctx, user.ID)
	if primary == nil {
		return fiber.NewError(400, "Your account has no primary email to send the test email to")
	}

	brand := star.Branding.Load()

	err = email.NewProvider(config).Send(
		primary.Email,
		"Test email from "+brand.Name,
		emails.Test,
		emails.TestData{
			Brand:       branding.ForEmail(brand, star.Config.ServerURL),
			Name:        user.Name,
			Provider:    config.Describe(),
			SentAt:      emails.FormatTime(time.Now()),
			SettingsURL: emails.EmailSettingsURL(star.Config.ServerURL),
		},
	)

	if err != nil {
		slog.Warn("failed to send test email", "err", err, "provider", config.Provider)
		return fiber.NewError(502, "Failed to send the test email: "+err.Error())
	}

	return ctx.JSON(types.TestEmailResponse{
		Email: primary.Email,
	})
}

func testCaptchaSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.TestCaptchaRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	applyCaptchaSettings(settings, &data.UpdateSettingsRequest)

	config := settings.GetCaptchaConfig()
	if !config.IsEnabled() {
		return fiber.NewError(400, "Choose a captcha provider before testing it")
	}

	if err := config.Validate(); err != nil {
		return err
	}

	if err := captcha.NewVerifier(config).Check(data.Token); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func applyEmailSettings(settings *database.Settings, request *types.UpdateSettingsRequest) {
	if request.EmailProvider != nil {
		settings.EmailProvider = email.Provider(*request.EmailProvider)
	}

	if request.EmailSender != nil {
		settings.EmailSender = strings.TrimSpace(*request.EmailSender)
	}

	if request.SMTPHost != nil {
		settings.SMTPHost = strings.TrimSpace(*request.SMTPHost)
	}

	if request.SMTPPort != nil {
		settings.SMTPPort = *request.SMTPPort
	}

	if request.SMTPUsername != nil {
		settings.SMTPUsername = strings.TrimSpace(*request.SMTPUsername)
	}

	if request.SMTPPassword != nil {
		settings.SMTPPassword = *request.SMTPPassword
	}

	if request.ResendAPIKey != nil {
		settings.ResendAPIKey = strings.TrimSpace(*request.ResendAPIKey)
	}
}

func applyCaptchaSettings(settings *database.Settings, request *types.UpdateSettingsRequest) {
	if request.CaptchaProvider != nil {
		settings.CaptchaProvider = captcha.Provider(*request.CaptchaProvider)
	}

	if request.TurnstileSiteKey != nil {
		settings.TurnstileSiteKey = strings.TrimSpace(*request.TurnstileSiteKey)
	}

	if request.TurnstileSecretKey != nil {
		settings.TurnstileSecretKey = strings.TrimSpace(*request.TurnstileSecretKey)
	}
}

func changedSecrets(request *types.UpdateSettingsRequest) []string {
	changed := []string{}

	if request.SMTPPassword != nil {
		changed = append(changed, "smtp_password")
	}

	if request.ResendAPIKey != nil {
		changed = append(changed, "resend_api_key")
	}

	if request.TurnstileSecretKey != nil {
		changed = append(changed, "turnstile_secret_key")
	}

	return changed
}
