package types

import "github.com/tubblelive/star/internal/password"

type UpdatePasswordSettingsRequest struct {
	MinLength  *int                `json:"min_length" validate:"omitempty,min=8,max=128"`
	MaxLength  *int                `json:"max_length" validate:"omitempty,min=8,max=512"`
	MinSymbols *int                `json:"min_symbols" validate:"omitempty,min=0,max=32"`
	MinUpper   *int                `json:"min_upper" validate:"omitempty,min=0,max=32"`
	MinLower   *int                `json:"min_lower" validate:"omitempty,min=0,max=32"`
	Algorithm  *password.Algorithm `json:"algorithm" validate:"omitempty,oneof=argon2 bcrypt"`
}

type UpdateSettingsRequest struct {
	AllowRegistration      *bool                          `json:"allow_registration" validate:"omitempty"`
	AllowLogins            *bool                          `json:"allow_logins" validate:"omitempty"`
	EmailCodeLength        *int                           `json:"email_code_length" validate:"omitempty,min=6,max=12"`
	EmailVerification      *bool                          `json:"email_verification" validate:"omitempty"`
	EmailDomainMode        *string                        `json:"email_domain_mode" validate:"omitempty,oneof=disabled whitelist blacklist"`
	EmailDomains           []string                       `json:"email_domains" validate:"omitempty,max=1000,dive,max=253"`
	BlockDisposableEmails  *bool                          `json:"block_disposable_emails" validate:"omitempty"`
	TrackBirthdays         *bool                          `json:"track_birthdays" validate:"omitempty"`
	MinimumAge             *int                           `json:"minimum_age" validate:"omitempty,min=0,max=120"`
	PasswordSettings       *UpdatePasswordSettingsRequest `json:"password_settings" validate:"omitempty"`
	EmailProvider          *string                        `json:"email_provider" validate:"omitempty,oneof=disabled smtp resend"`
	EmailSender            *string                        `json:"email_sender" validate:"omitempty,max=320"`
	SMTPHost               *string                        `json:"smtp_host" validate:"omitempty,max=255"`
	SMTPPort               *int                           `json:"smtp_port" validate:"omitempty,min=1,max=65535"`
	SMTPUsername           *string                        `json:"smtp_username" validate:"omitempty,max=255"`
	SMTPPassword           *string                        `json:"smtp_password" validate:"omitempty,max=1024"`
	ResendAPIKey           *string                        `json:"resend_api_key" validate:"omitempty,max=1024"`
	CaptchaProvider        *string                        `json:"captcha_provider" validate:"omitempty,oneof=disabled turnstile"`
	TurnstileSiteKey       *string                        `json:"turnstile_site_key" validate:"omitempty,max=255"`
	TurnstileSecretKey     *string                        `json:"turnstile_secret_key" validate:"omitempty,max=1024"`
	LockoutThreshold       *int                           `json:"lockout_threshold" validate:"omitempty,min=3,max=100"`
	LockoutWindowMinutes   *int                           `json:"lockout_window_minutes" validate:"omitempty,min=1,max=1440"`
	LockoutDurationMinutes *int                           `json:"lockout_duration_minutes" validate:"omitempty,min=1,max=1440"`
}

type TestCaptchaRequest struct {
	UpdateSettingsRequest
	Token string `json:"token" validate:"required,max=2048"`
}

type TestEmailResponse struct {
	Email string `json:"email"`
}
