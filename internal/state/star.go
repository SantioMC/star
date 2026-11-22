package state

import (
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/geolocation"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/revocation"
)

type StarServer struct {
	// Server data
	Database   *database.Database
	Config     *config.Config
	GeoService geolocation.Service
	WebAuthn   *webauthn.WebAuthn
	StartedAt  time.Time

	// Cached data
	Settings    atomic.Pointer[database.Settings]
	Branding    atomic.Pointer[database.Branding]
	DefaultApp  atomic.Pointer[database.Application]
	SigningKeys atomic.Pointer[[]database.SigningKey]
	Revocations *revocation.List
}

func (s *StarServer) EmailConfig() email.Config {
	settings := s.Settings.Load()
	if settings == nil {
		return email.Config{Provider: email.ProviderDisabled}
	}

	return settings.GetEmailConfig()
}

func (s *StarServer) Email() email.EmailProvider {
	return email.NewProvider(s.EmailConfig())
}

func (s *StarServer) CaptchaConfig() captcha.Config {
	settings := s.Settings.Load()
	if settings == nil {
		return captcha.Config{Provider: captcha.ProviderDisabled}
	}

	return settings.GetCaptchaConfig()
}

func (s *StarServer) Captcha() captcha.Verifier {
	return captcha.NewVerifier(s.CaptchaConfig())
}

func (s *StarServer) PasswordHasher() cryptography.PasswordHasher {
	settings := s.Settings.Load()
	if settings == nil {
		return cryptography.NewPasswordHasher(password.DefaultSettings.Algorithm)
	}

	return settings.GetPasswordHasher()
}

func (s *StarServer) ClientIP(ctx fiber.Ctx) string {
	remote, _ := netip.AddrFromSlice(ctx.RequestCtx().RemoteIP())
	return s.Config.Proxy.ClientIP(remote, ctx.Get(s.Config.Proxy.Header))
}
