package server

import "github.com/gofiber/fiber/v3"

func healthRouter(router fiber.Router) {
	router.Get("/", healthHandler)
}

func healthHandler(ctx fiber.Ctx) error {
	return ctx.SendString("OK")
}
