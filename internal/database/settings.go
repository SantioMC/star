package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

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
	AuditRetentionDays    *int                      `json:"audit_retention_days" db:"audit_retention_days"`
	AuditAnonymize        bool                      `json:"audit_anonymize" db:"audit_anonymize"`
	PasswordSettings      string                    `json:"password_settings" db:"password_settings"`
}

func (s *Settings) GetEmailDomainPolicy() email.DomainPolicy {
	return email.DomainPolicy{
		Mode:            s.EmailDomainMode,
		Domains:         s.EmailDomains.Data,
		BlockDisposable: s.BlockDisposableEmails,
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

	return &settings, nil
}

func (db *Database) UpdateSettings(ctx context.Context, settings *Settings) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE settings
		SET allow_registration = $2, allow_logins = $3, password_settings = $4, email_code_length = $5,
			audit_retention_days = $6, audit_anonymize = $7, email_verification = $8, email_domain_mode = $9,
			email_domains = $10, block_disposable_emails = $11
		WHERE id = $1
		`,
		settings.ID,
		settings.AllowRegistration,
		settings.AllowLogins,
		settings.PasswordSettings,
		settings.EmailCodeLength,
		settings.AuditRetentionDays,
		settings.AuditAnonymize,
		settings.EmailVerification,
		settings.EmailDomainMode,
		settings.EmailDomains,
		settings.BlockDisposableEmails,
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
