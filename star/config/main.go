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
)

type Config struct {
	Database        string
	Url             string
	Email           string
	Origins         []string
	TrustProxy      bool
	Port            int
	AccessToken     [32]byte
	CaptchaVerifier *captcha.Verifier
}

func LoadFromEnv() (*Config, error) {
	config := Config{
		Database:    fromEnv("DATABASE", "sqlite://./data.db"),
		Url:         fromEnv("URL", ""),
		Email:       fromEnv("EMAIL", ""),
		Origins:     []string{"*"},
		TrustProxy:  false,
		AccessToken: [32]byte{},
		Port:        8080,
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
	config.CaptchaVerifier = &captcha

	if config.CaptchaVerifier == nil {
		slog.Warn("running star without a captcha configured")
	}

	return &config, nil
}

func loadCaptchaFromEnv() captcha.Verifier {
	captcha := fromEnv("CAPTCHA", "")
	if captcha == "" {
		return nil
	}

	switch strings.ToLower(captcha) {
	case "cloudflare", "turnstile":
		siteKey := fromEnv("TURNSTILE_SITE_KEY", "")
		if siteKey == "" {
			slog.Error("expected TURNSTILE_SITE_KEY to be set", "captcha", captcha)
			os.Exit(1)
		}

		captchaToken := fromEnv("TURNSTILE_TOKEN", "")
		if captchaToken == "" {
			slog.Error("expected TURNSTILE_TOKEN to be set", "captcha", captcha)
			os.Exit(1)
		}

		turnstile := turnstile.NewTurnstileValidator(siteKey, captchaToken)
		return &turnstile
	default:
		slog.Warn("unknown captcha software", "captcha", captcha)
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
