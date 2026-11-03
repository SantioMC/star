package request

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/state"
)

func SecureCookies(star *state.StarServer) bool {
	return strings.HasPrefix(star.Config.ServerURL, "https://")
}

func SetCookie(ctx fiber.Ctx, star *state.StarServer, cookie fiber.Cookie) {
	cookie.HTTPOnly = true
	cookie.Secure = SecureCookies(star)

	ctx.Cookie(&cookie)
}
