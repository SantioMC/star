package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/tubblelive/star/internal/server/oauth"
)

func oauthRouter(router fiber.Router) {
	tokenLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: time.Minute,
		LimitReached: func(ctx fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	})

	router.Get("/authorize", authorizeHandler)
	router.Get("/application/:client_id", applicationHandler)
	router.Post("/token", tokenLimiter, tokenHandler)
	router.Post("/revoke", tokenLimiter, revokeHandler)
}

func authorizeHandler(ctx fiber.Ctx) error {
	return oauth.Authorize(ctx)
}

func applicationHandler(ctx fiber.Ctx) error {
	return oauth.Application(ctx)
}

func tokenHandler(ctx fiber.Ctx) error {
	return oauth.Token(ctx)
}

func revokeHandler(ctx fiber.Ctx) error {
	return oauth.Revoke(ctx)
}
