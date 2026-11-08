package captcha

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/captcha/turnstile"
)

type Provider string

const (
	ProviderDisabled  Provider = "disabled"
	ProviderTurnstile Provider = "turnstile"
)

type Config struct {
	Provider           Provider
	TurnstileSiteKey   string
	TurnstileSecretKey string
}

func (c Config) IsEnabled() bool {
	return c.Provider != ProviderDisabled
}

func (c Config) Validate() error {
	switch c.Provider {
	case ProviderDisabled:
		return nil
	case ProviderTurnstile:
		if c.TurnstileSiteKey == "" || c.TurnstileSecretKey == "" {
			return fiber.NewError(400, "Turnstile requires both a site key and a secret key")
		}

		return nil
	default:
		return fiber.NewError(400, "Unknown captcha provider")
	}
}

func (c Config) CheckSecret() error {
	err := NewVerifier(c).Check("secret-check")
	if errors.Is(err, turnstile.ErrInvalidSecret) || errors.Is(err, turnstile.ErrUnavailable) {
		return err
	}

	return nil
}

func NewVerifier(c Config) Verifier {
	switch c.Provider {
	case ProviderTurnstile:
		return turnstile.NewTurnstileValidator(c.TurnstileSiteKey, c.TurnstileSecretKey)
	default:
		return &NoopVerifier{}
	}
}
