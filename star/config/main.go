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
	Origins         []string
	TrustProxy      bool
	Port            int
	AccessToken     [32]byte
	CaptchaVerifier captcha.Verifier
}

func LoadFromEnv() (*Config, error) {
	config := Config{
		Database:    fromEnv("DATABASE", "sqlite://./data.db"),
		Url:         fromEnv("URL", ""),
		Origins:     []string{"*"},
		TrustProxy:  false,
		AccessToken: [32]byte{},
		Port:        8080,
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
		// There should be no reason for us to ever ask for it again, as bonus we only store the hash in memory
		os.Unsetenv("ACCESS_TOKEN") 
	}

	config.Origins = strings.Split(fromEnv("ORIGINS", "*"), ",")
	config.TrustProxy = fromEnv("TRUST_PROXY", "false") == "true"
	config.CaptchaVerifier = loadCaptchaFromEnv()

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
	case "cloudflare":
		captchaToken := fromEnv("CLOUDFLARE_TOKEN", "")
		if captchaToken == "" {
			slog.Error("expected CLOUDFLARE_TOKEN to be set when CAPTCHA is set to 'cloudflare'")
			os.Exit(1)
		}

		turnstile := turnstile.NewTurnstileValidator(captchaToken)
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
