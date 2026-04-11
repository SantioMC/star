package config

type Config struct {
	ServerURL   string   `env:"SERVER_URL"`
	Port        int      `env:"PORT,8080"`
	DatabaseURL string   `env:"DATABASE_URL"`
	Origins     []string `env:"ORIGIN,"`
	TrustProxy  bool     `env:"TRUST_PROXY,false"`
	ProxyHeader string   `env:"PROXY_HEADER,X-Forwarded-For"`
	Captcha     CaptchaConfig
	Email       EmailConfig
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	err := loadConfig(config)

	return config, err
}
