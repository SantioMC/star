package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/oauth"
)

func oauthRouter(router fiber.Router) {
	tokenLimiter := middleware.RateLimit(60)

	router.Get("/authorize", authorizeHandler)
	router.Get("/application/:client_id", tokenLimiter, applicationHandler)
	router.Post("/token", tokenLimiter, tokenHandler)
	router.Post("/revoke", tokenLimiter, revokeHandler)
	router.Post("/introspect", middleware.RateLimit(300), introspectHandler)
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

func introspectHandler(ctx fiber.Ctx) error {
	return oauth.Introspect(ctx)
}
