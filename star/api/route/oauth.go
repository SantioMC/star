package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/api/route/oauth"
)

func OAuthRouter(r fiber.Router) {
	rateLimiter := limiter.New(helper.HeavyLimiter)
	r.Post("token", rateLimiter, oauth.HandleOidcToken)
}