package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/tubblelive/star/internal/birthday"
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/password"
	"github.com/vinovest/sqlx"
)

type Settings struct {
	ID                    int                       `json:"id"`
	AllowRegistration     bool                      `json:"allow_registration" db:"allow_registration"`
	AllowLogins           bool                      `json:"allow_logins" db:"allow_logins"`
	EmailCodeLength       int                       `json:"email_code_length" db:"email_code_length"`
	EmailVerification     bool                      `json:"email_verification" db:"email_verification"`
	EmailDomainMode       email.DomainMode          `json:"email_domain_mode" db:"email_domain_mode"`
	EmailDomains          models.JSONData[[]string] `json:"email_domains" db:"email_domains"`
	BlockDisposableEmails bool                      `json:"block_disposable_emails" db:"block_disposable_emails"`
	TrackBirthdays        bool                      `json:"track_birthdays" db:"track_birthdays"`
	MinimumAge            int                       `json:"minimum_age" db:"minimum_age"`
	AuditRetentionDays    *int                      `json:"audit_retention_days" db:"audit_retention_days"`
	AuditAnonymize        bool                      `json:"audit_anonymize" db:"audit_anonymize"`
	PasswordSettings      string                    `json:"password_settings" db:"password_settings"`
	EmailProvider         email.Provider            `json:"email_provider" db:"email_provider"`
	EmailSender           string                    `json:"email_sender" db:"email_sender"`
	SMTPHost              string                    `json:"smtp_host" db:"smtp_host"`
	SMTPPort              int                       `json:"smtp_port" db:"smtp_port"`
	SMTPUsername          string                    `json:"smtp_username" db:"smtp_username"`
	SMTPPassword          string                    `json:"-" db:"smtp_password"`
	ResendAPIKey          string                    `json:"-" db:"resend_api_key"`
	CaptchaProvider       captcha.Provider          `json:"captcha_provider" db:"captcha_provider"`
	TurnstileSiteKey      string                    `json:"turnstile_site_key" db:"turnstile_site_key"`
	TurnstileSecretKey    string                    `json:"-" db:"turnstile_secret_key"`
}

func (s *Settings) GetEmailConfig() email.Config {
	return email.Config{
		Provider:     s.EmailProvider,
		Sender:       s.EmailSender,
		SMTPHost:     s.SMTPHost,
		SMTPPort:     s.SMTPPort,
		SMTPUsername: s.SMTPUsername,
		SMTPPassword: s.SMTPPassword,
		ResendAPIKey: s.ResendAPIKey,
	}
}

func (s *Settings) GetCaptchaConfig() captcha.Config {
	return captcha.Config{
		Provider:           s.CaptchaProvider,
		TurnstileSiteKey:   s.TurnstileSiteKey,
		TurnstileSecretKey: s.TurnstileSecretKey,
	}
}

func (s *Settings) secrets() []*string {
	return []*string{&s.SMTPPassword, &s.ResendAPIKey, &s.TurnstileSecretKey}
}

func (s *Settings) transformSecrets(key [32]byte, transform func(string, [32]byte) (string, error)) error {
	for _, secret := range s.secrets() {
		if *secret == "" {
			continue
		}

		transformed, err := transform(*secret, key)
		if err != nil {
			return err
		}

		*secret = transformed
	}

	return nil
}

func (s *Settings) GetEmailDomainPolicy() email.DomainPolicy {
	return email.DomainPolicy{
		Mode:            s.EmailDomainMode,
		Domains:         s.EmailDomains.Data,
		BlockDisposable: s.BlockDisposableEmails,
	}
}

func (s *Settings) GetAgePolicy() birthday.Policy {
	return birthday.Policy{
		Tracked:    s.TrackBirthdays,
		MinimumAge: s.MinimumAge,
	}
}

func (s *Settings) GetPasswordSettings() *password.PasswordSetttings {
	var settings password.PasswordSetttings

	if err := json.Unmarshal([]byte(s.PasswordSettings), &settings); err != nil {
		slog.Error("failed to unmarshal password settings", "err", err)
		return &password.DefaultSettings
	}

	return &settings
}

func (db *Database) GetSettings(ctx context.Context) (*Settings, error) {
	settings, err := sqlx.OneContext[Settings](
		ctx,
		db.DB(),
		`SELECT * FROM settings LIMIT 1`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		slog.Error("failed to get settings, this is a fatal error", "err", err)
		return nil, err
	}

	if err := settings.transformSecrets(db.config.AppToken, cryptography.DecryptAES); err != nil {
		slog.Error("failed to decrypt settings secrets", "err", err)
		return nil, err
	}

	return &settings, nil
}

func (db *Database) UpdateSettings(ctx context.Context, settings *Settings) error {
	encrypted := *settings
	if err := encrypted.transformSecrets(db.config.AppToken, cryptography.EncryptAES); err != nil {
		return err
	}

	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE settings
		SET allow_registration = $2, allow_logins = $3, password_settings = $4, email_code_length = $5,
			audit_retention_days = $6, audit_anonymize = $7, email_verification = $8, email_domain_mode = $9,
			email_domains = $10, block_disposable_emails = $11, minimum_age = $12,
			track_birthdays = $13, email_provider = $14, email_sender = $15, smtp_host = $16,
			smtp_port = $17, smtp_username = $18, smtp_password = $19, resend_api_key = $20,
			captcha_provider = $21, turnstile_site_key = $22, turnstile_secret_key = $23
		WHERE id = $1
		`,
		encrypted.ID,
		encrypted.AllowRegistration,
		encrypted.AllowLogins,
		encrypted.PasswordSettings,
		encrypted.EmailCodeLength,
		encrypted.AuditRetentionDays,
		encrypted.AuditAnonymize,
		encrypted.EmailVerification,
		encrypted.EmailDomainMode,
		encrypted.EmailDomains,
		encrypted.BlockDisposableEmails,
		encrypted.MinimumAge,
		encrypted.TrackBirthdays,
		encrypted.EmailProvider,
		encrypted.EmailSender,
		encrypted.SMTPHost,
		encrypted.SMTPPort,
		encrypted.SMTPUsername,
		encrypted.SMTPPassword,
		encrypted.ResendAPIKey,
		encrypted.CaptchaProvider,
		encrypted.TurnstileSiteKey,
		encrypted.TurnstileSecretKey,
	)

	return err
}

func (db *Database) SaveDefaultSettings(ctx context.Context) error {
	passwordSettings, err := json.Marshal(password.DefaultSettings)
	if err != nil {
		slog.Error("failed to marshal default password settings", "err", err)
		return err
	}

	_, err = db.DB().ExecContext(
		ctx,
		`
		INSERT INTO settings (id, allow_registration, allow_logins, password_settings, email_code_length)
		VALUES (1, $1, $2, $3, $4)
		`,
		true,
		true,
		passwordSettings,
		8,
	)

	return err
}

func (db *Database) SettingsExists(ctx context.Context) bool {
	exists, err := sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`SELECT EXISTS(SELECT 1 FROM settings LIMIT 1)`,
	)

	if err != nil {
		slog.Error("failed to check if settings exist", "err", err)
		return false
	}

	return exists
}
