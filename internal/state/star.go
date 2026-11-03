package state

import (
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/geolocation"
	"github.com/tubblelive/star/internal/revocation"
)

type StarServer struct {
	// Server data
	Database   *database.Database
	Config     *config.Config
	GeoService geolocation.Service
	Captcha    captcha.Verifier
	Email      email.EmailProvider
	WebAuthn   *webauthn.WebAuthn
	StartedAt  time.Time

	// Cached data
	Settings    atomic.Pointer[database.Settings]
	DefaultApp  atomic.Pointer[database.Application]
	SigningKeys atomic.Pointer[[]database.SigningKey]
	Revocations *revocation.List
}

func (s *StarServer) ClientIP(ctx fiber.Ctx) string {
	remote, _ := netip.AddrFromSlice(ctx.RequestCtx().RemoteIP())
	return s.Config.Proxy.ClientIP(remote, ctx.Get(s.Config.Proxy.Header))
}
