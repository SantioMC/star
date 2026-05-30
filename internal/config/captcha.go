package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/captcha/turnstile"
)

type CaptchaConfig struct {
	Type      string `env:"CAPTCHA,disabled"`
	Turnstile TurnstileConfig
}

type TurnstileConfig struct {
	Token   *string `env:"TURNSTILE_TOKEN"`
	SiteKey *string `env:"TURNSTILE_SITE_KEY"`
}

func (c *CaptchaConfig) IsEnabled() bool {
	return c.Type != "disabled"
}

func (c *CaptchaConfig) CreateVerifier() captcha.Verifier {
	switch strings.ToLower(c.Type) {
	case "turnstile":
		if c.Turnstile.Token == nil || c.Turnstile.SiteKey == nil {
			slog.Error("turnstile captcha was specified, required TURNSTILE_TOKEN and TURNSTILE_SITE_KEY environment variables not set")
			os.Exit(1)
		}

		validator := turnstile.NewTurnstileValidator(*c.Turnstile.SiteKey, *c.Turnstile.Token)
		return &validator
	default:
		return &captcha.NoopVerifier{}
	}
}
