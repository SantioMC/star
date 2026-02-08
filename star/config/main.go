package config

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/tubblelive/star/captcha"
	"github.com/tubblelive/star/captcha/turnstile"
	"github.com/tubblelive/star/email"
	"github.com/tubblelive/star/email/provider"
)

type Config struct {
	Database        string
	Url             string
	SecurityEmail   string
	Origins         []string
	TrustProxy      bool
	Port            int
	AccessToken     [32]byte
	CaptchaVerifier captcha.Verifier
	EmailProvider   email.EmailProvider
}

func LoadFromEnv() (*Config, error) {
	config := Config{
		Database:      fromEnv("DATABASE", "sqlite://./data.db"),
		Url:           fromEnv("URL", ""),
		SecurityEmail: fromEnv("SECURITY_EMAIL", ""),
		Origins:       []string{"*"},
		TrustProxy:    false,
		AccessToken:   [32]byte{},
		Port:          8080,
	}

	if config.Url == "" {
		return nil, fmt.Errorf("URL env is required")
	}

	var err error
	if config.Port, err = strconv.Atoi(fromEnv("PORT", "8080")); err != nil {
		return nil, err
	}

	accessToken := fromEnv("ACCESS_TOKEN", "")
	if accessToken == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN env is required")
	} else {
		config.AccessToken = sha256.Sum256([]byte(accessToken))
		os.Unsetenv("ACCESS_TOKEN")
	}

	origins := os.Getenv("ORIGINS")
	if origins == "" {
		if config.Url != "" {
			index := strings.Index(config.Url, "://") + 3
			config.Origins = []string{config.Url[index:]}
		} else {
			config.Origins = []string{"*"}
		}
	} else {
		config.Origins = strings.Split(origins, ",")
	}

	config.TrustProxy = fromEnv("TRUST_PROXY", "false") == "true"

	captcha := loadCaptchaFromEnv()
	config.CaptchaVerifier = captcha
	if config.CaptchaVerifier == nil {
		slog.Warn("running star without a captcha configured")
	}

	email := loadEmailFromEnv()
	config.EmailProvider = email
	if config.EmailProvider == nil {
		slog.Warn("running star without an email provider configured")
	}

	return &config, nil
}

func getOrFail(env string, def string) string {
	value := fromEnv(env, def)
	if value == "" {
		slog.Error("expected " + env + " to be set")
		os.Exit(1)
	}

	return value
}

func loadCaptchaFromEnv() captcha.Verifier {
	captcha := fromEnv("CAPTCHA", "")
	if captcha == "" {
		return nil
	}

	switch strings.ToLower(captcha) {
	case "cloudflare", "turnstile":
		siteKey := getOrFail("TURNSTILE_SITE_KEY", "")
		captchaToken := getOrFail("TURNSTILE_TOKEN", "")

		turnstile := turnstile.NewTurnstileValidator(siteKey, captchaToken)
		return &turnstile
	default:
		slog.Warn("unknown captcha software", "captcha", captcha)
	}

	return nil
}

func loadEmailFromEnv() email.EmailProvider {
	email := fromEnv("EMAIL_PROVIDER", "")
	if email == "" {
		return nil
	}

	sender := getOrFail("EMAIL_SENDER", "")

	switch strings.ToLower(email) {
	case "resend":
		apiKey := getOrFail("RESEND_API_KEY", "")
		return provider.NewResendProvider(sender, apiKey)
	case "smtp":
		host := getOrFail("SMTP_HOST", "")
		rawPort := getOrFail("SMTP_PORT", "465")
		username := getOrFail("SMTP_USERNAME", "")
		password := getOrFail("SMTP_PASSWORD", "")

		port, err := strconv.Atoi(rawPort)
		if err != nil {
			slog.Error("expected SMTP_PORT to be a number", "email", email, "port", rawPort)
			os.Exit(1)
		}

		return provider.NewSMTPProvider(
			sender,
			host,
			port,
			username,
			password,
		)
	default:
		slog.Warn("unknown email provider", "email", email)
	}

	return nil
}

func fromEnv(env string, def string) string {
	if value, ok := os.LookupEnv(env); ok {
		if value == "" {
			return def
		}

		return value
	}

	return def
}
