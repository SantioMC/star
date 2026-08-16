package config

import (
	"crypto/sha256"
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	AppToken    [32]byte `env:"-"`
	ServerURL   string   `env:"SERVER_URL"`
	Port        int      `env:"PORT,8080"`
	DatabaseURL string   `env:"DATABASE_URL"`
	Origins     []string `env:"ORIGIN,"`
	TrustProxy  bool     `env:"TRUST_PROXY,false"`
	ProxyHeader string   `env:"PROXY_HEADER,X-Forwarded-For"`
	Captcha     CaptchaConfig
	Email       EmailConfig
	Geolocation GeoserviceConfig
}

func LoadConfig() (*Config, error) {
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		return nil, errors.New("APP_TOKEN is required")
	}

	if len(appToken) < 32 {
		// Warn if given a poorly chosen appToken but still allow it through in case this is a dev environment
		slog.Warn("APP_TOKEN is low entropy, please use a stronger token in production")
	}

	config := &Config{}
	err := loadConfig(config)
	config.AppToken = sha256.Sum256([]byte(appToken))

	return config, err
}
