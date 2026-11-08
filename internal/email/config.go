package email

import (
	"fmt"
	"net"
	"net/mail"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type Provider string

const (
	ProviderDisabled Provider = "disabled"
	ProviderSMTP     Provider = "smtp"
	ProviderResend   Provider = "resend"
)

type Config struct {
	Provider     Provider
	Sender       string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	ResendAPIKey string
}

func (c Config) IsEnabled() bool {
	return c.Provider != ProviderDisabled
}

func (c Config) Validate() error {
	if !c.IsEnabled() {
		return nil
	}

	if c.Sender == "" {
		return fiber.NewError(400, "Sending emails requires a sender address")
	}

	if _, err := mail.ParseAddress(c.Sender); err != nil {
		return fiber.NewError(400, "The sender must be a valid address, such as \"Star <noreply@example.com>\"")
	}

	switch c.Provider {
	case ProviderSMTP:
		if c.SMTPHost == "" {
			return fiber.NewError(400, "SMTP requires a host")
		}

		if c.SMTPPort < 1 || c.SMTPPort > 65535 {
			return fiber.NewError(400, "The SMTP port must be between 1 and 65535")
		}
	case ProviderResend:
		if c.ResendAPIKey == "" {
			return fiber.NewError(400, "Resend requires an API key")
		}
	default:
		return fiber.NewError(400, "Unknown email provider")
	}

	return nil
}

func (c Config) Describe() string {
	switch c.Provider {
	case ProviderSMTP:
		return fmt.Sprintf("SMTP (%s)", net.JoinHostPort(c.SMTPHost, strconv.Itoa(c.SMTPPort)))
	case ProviderResend:
		return "Resend"
	default:
		return "nothing"
	}
}

func NewProvider(c Config) EmailProvider {
	switch c.Provider {
	case ProviderSMTP:
		return NewSMTPProvider(c.Sender, c.SMTPHost, c.SMTPPort, c.SMTPUsername, c.SMTPPassword)
	case ProviderResend:
		return NewResendProvider(c.Sender, c.ResendAPIKey)
	default:
		return &NoopProvider{}
	}
}
