package api

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/star/api/route"
	"github.com/tubblelive/star/star/state"
)

func StartAPIServer(server *state.Server) *fiber.App {
	config := fiber.Config{
		ServerHeader:          "Star",
		AppName:               "Star Authorization Server",
		DisableStartupMessage: true,
	}

	app := fiber.New(config)

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowMethods:     "GET, POST, PATCH, DELETE, OPTIONS",
		AllowOriginsFunc: func(origin string) bool {
			for _, allowedOrigin := range server.Config.Origins {
				if origin == allowedOrigin {
					return true
				}
			}

			return false
		},
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	app.Use(compress.New())

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			switch data := err.(type) {
			case *fiber.Error:
				c.Status(data.Code).JSON(fiber.Map{
					"error": true,
					"status": data.Code,
					"message": data.Message,
				})
			default:
				slog.Error("failed handling request", "err", err)
				c.Status(500).JSON(fiber.Map{
					"error": true,
					"status": 500,
					"message": data.Error(),
				})
			}
		}

		return nil
	})

	// Route handlers
	app.Get("/healthz", route.Health)
	app.Route("/.well-known", route.WellKnownRouter)

	api := app.Group("/api")
	api.Route("/user", route.UserRouter)
	api.Route("/callback", route.LoginRouter)
	api.Route("/app", route.AppRouter)

	// Bind to port
	bind := "0.0.0.0:" + strconv.Itoa(server.Config.Port)
	slog.Info("starting api server", "bind", bind)

	if server.Config.Origins[0] == "*" {
		slog.Warn("origins is set to '*', you are vulnerable to security vulnerabilities")
	}

	err := app.Listen(bind)
	if err != nil {
		slog.Error("failed to start api server", "error", err)
	}

	return app
}
