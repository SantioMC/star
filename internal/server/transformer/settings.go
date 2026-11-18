package transformer

import (
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/password"
)

type AdminSettings struct {
	AllowRegistration      bool                       `json:"allow_registration"`
	AllowLogins            bool                       `json:"allow_logins"`
	EmailCodeLength        int                        `json:"email_code_length"`
	EmailVerification      bool                       `json:"email_verification"`
	EmailDomainMode        email.DomainMode           `json:"email_domain_mode"`
	EmailDomains           []string                   `json:"email_domains"`
	BlockDisposableEmails  bool                       `json:"block_disposable_emails"`
	TrackBirthdays         bool                       `json:"track_birthdays"`
	MinimumAge             int                        `json:"minimum_age"`
	PasswordSettings       password.PasswordSetttings `json:"password_settings"`
	EmailProvider          email.Provider             `json:"email_provider"`
	EmailSender            string                     `json:"email_sender"`
	SMTPHost               string                     `json:"smtp_host"`
	SMTPPort               int                        `json:"smtp_port"`
	SMTPUsername           string                     `json:"smtp_username"`
	HasSMTPPassword        bool                       `json:"has_smtp_password"`
	HasResendAPIKey        bool                       `json:"has_resend_api_key"`
	CaptchaProvider        captcha.Provider           `json:"captcha_provider"`
	TurnstileSiteKey       string                     `json:"turnstile_site_key"`
	HasTurnstileSecretKey  bool                       `json:"has_turnstile_secret_key"`
	LockoutThreshold       int                        `json:"lockout_threshold"`
	LockoutWindowMinutes   int                        `json:"lockout_window_minutes"`
	LockoutDurationMinutes int                        `json:"lockout_duration_minutes"`
}

func TransformSettings(settings database.Settings, passwordSettings password.PasswordSetttings) AdminSettings {
	return AdminSettings{
		AllowRegistration:      settings.AllowRegistration,
		AllowLogins:            settings.AllowLogins,
		EmailCodeLength:        settings.EmailCodeLength,
		EmailVerification:      settings.EmailVerification,
		EmailDomainMode:        settings.EmailDomainMode,
		EmailDomains:           settings.EmailDomains.Data,
		BlockDisposableEmails:  settings.BlockDisposableEmails,
		TrackBirthdays:         settings.TrackBirthdays,
		MinimumAge:             settings.MinimumAge,
		PasswordSettings:       passwordSettings,
		EmailProvider:          settings.EmailProvider,
		EmailSender:            settings.EmailSender,
		SMTPHost:               settings.SMTPHost,
		SMTPPort:               settings.SMTPPort,
		SMTPUsername:           settings.SMTPUsername,
		HasSMTPPassword:        settings.SMTPPassword != "",
		HasResendAPIKey:        settings.ResendAPIKey != "",
		CaptchaProvider:        settings.CaptchaProvider,
		TurnstileSiteKey:       settings.TurnstileSiteKey,
		HasTurnstileSecretKey:  settings.TurnstileSecretKey != "",
		LockoutThreshold:       settings.LockoutThreshold,
		LockoutWindowMinutes:   settings.LockoutWindowMinutes,
		LockoutDurationMinutes: settings.LockoutDurationMinutes,
	}
}

type AuditSettings struct {
	RetentionDays *int `json:"retention_days"`
	Anonymize     bool `json:"anonymize"`
}

func TransformAuditSettings(settings database.Settings) AuditSettings {
	return AuditSettings{
		RetentionDays: settings.AuditRetentionDays,
		Anonymize:     settings.AuditAnonymize,
	}
}
