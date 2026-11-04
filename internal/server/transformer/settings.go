package transformer

import (
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/password"
)

type AdminSettings struct {
	AllowRegistration bool                       `json:"allow_registration"`
	AllowLogins       bool                       `json:"allow_logins"`
	EmailCodeLength   int                        `json:"email_code_length"`
	PasswordSettings  password.PasswordSetttings `json:"password_settings"`
}

func TransformSettings(settings database.Settings, passwordSettings password.PasswordSetttings) AdminSettings {
	return AdminSettings{
		AllowRegistration: settings.AllowRegistration,
		AllowLogins:       settings.AllowLogins,
		EmailCodeLength:   settings.EmailCodeLength,
		PasswordSettings:  passwordSettings,
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
