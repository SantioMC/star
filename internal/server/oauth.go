package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/oauth"
)

func oauthRouter(router fiber.Router) {
	router.Get("/authorize", authorizeHandler)
}

func authorizeHandler(ctx fiber.Ctx) error {
	return oauth.Authorize(ctx)
}
