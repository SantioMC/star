package config

import (
	"log/slog"
	"os"

	"github.com/tubblelive/star/internal/email"
)

type EmailConfig struct {
	Type   string  `env:"EMAIL,disabled"`
	Sender *string `env:"EMAIL_SENDER"`
	Resend ResendConfig
	SMTP   SMTPConfig
}

type ResendConfig struct {
	APIKey *string `env:"RESEND_API_KEY"`
}

type SMTPConfig struct {
	Host     *string `env:"SMTP_HOST"`
	Port     *int    `env:"SMTP_PORT"`
	Username *string `env:"SMTP_USERNAME"`
	Password *string `env:"SMTP_PASSWORD"`
}

func (c *EmailConfig) IsEnabled() bool {
	return c.Type != "disabled"
}

func (c *EmailConfig) CreateSender() email.EmailProvider {
	if c.Sender == nil && c.IsEnabled() {
		slog.Error("email was specified, required EMAIL_SENDER environment variable not set")
		os.Exit(1)
	}

	switch c.Type {
	case "resend":
		if c.Resend.APIKey == nil {
			slog.Error("resend email was specified, required RESEND_API_KEY environment variable not set")
			os.Exit(1)
		}

		return email.NewResendProvider(*c.Sender, *c.Resend.APIKey)
	case "smtp":
		if c.SMTP.Host == nil || c.SMTP.Port == nil || c.SMTP.Username == nil || c.SMTP.Password == nil {
			slog.Error("smtp email was specified, required SMTP_HOST, SMTP_PORT, SMTP_USERNAME, and SMTP_PASSWORD environment variables not set")
			os.Exit(1)
		}

		return email.NewSMTPProvider(*c.Sender, *c.SMTP.Host, *c.SMTP.Port, *c.SMTP.Username, *c.SMTP.Password)
	default:
		return &email.NoopProvider{}
	}
}
