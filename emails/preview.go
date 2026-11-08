package emails

import (
	"html/template"
	"time"
)

func Sample(name string, brand Branding, serverURL string) (*template.Template, any, bool) {
	switch name {
	case "verification":
		return Verification, VerificationData{
			Brand:           brand,
			Name:            "jane",
			VerificationURL: VerificationURL(serverURL, "jane@example.com", "K7Q2M9XD"),
			Code:            "K7Q2M9XD",
		}, true
	case "password-reset":
		return PasswordReset, PasswordResetData{
			Brand:     brand,
			Name:      "jane",
			ResetURL:  PasswordResetURL(serverURL, "preview"),
			IPAddress: "203.0.113.42",
		}, true
	case "password-changed":
		return PasswordChanged, PasswordChangedData{
			Brand:       brand,
			Name:        "jane",
			ChangedAt:   FormatTime(time.Now()),
			IPAddress:   "203.0.113.42",
			RecoveryURL: RecoveryURL(serverURL),
		}, true
	case "invite":
		return Invite, InviteData{
			Brand:     brand,
			Name:      "jane",
			Inviter:   "admin",
			InviteURL: InviteURL(serverURL, "preview"),
			ExpiresAt: FormatTime(time.Now().Add(time.Hour * 72)),
		}, true
	case "test":
		return Test, TestData{
			Brand:       brand,
			Name:        "jane",
			Provider:    "SMTP (smtp.example.com:587)",
			SentAt:      FormatTime(time.Now()),
			SettingsURL: EmailSettingsURL(serverURL),
		}, true
	}

	return nil, nil, false
}
