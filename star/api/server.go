package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/frontend"
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

	app.Use(compress.New())

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		uri := string(c.Request().URI().RequestURI())
		request := c.Method() + " " + uri

		if err != nil {
			switch data := err.(type) {
			case *fiber.Error:
				slog.Error(strconv.Itoa(data.Code) + " " + request, "err", err)
				c.Status(data.Code).JSON(fiber.Map{
					"error": true,
					"status": data.Code,
					"message": data.Message,
				})
			default:
				slog.Error("500 " + request, "params", c.AllParams(), "err", err)
				c.Status(500).JSON(fiber.Map{
					"error": true,
					"status": 500,
					"message": data.Error(),
				})
			}
		} else if strings.HasPrefix(uri, "/api") {
			slog.Info(strconv.Itoa(c.Response().StatusCode()) + " " + request)
		}

		return nil
	})

	// Route handlers

	// Global
	global := app.Group("/", cors.New(cors.Config{
		AllowCredentials: false,
	}))

	global.Get("/healthz", route.Health)
	global.Route("/.well-known", route.WellKnownRouter)

	// API
	api := app.Group("/api", limiter.New(limiter.Config{
		Max:        300,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))
	
	api.Route("/", route.LoginRouter)
	api.Route("/user", route.UserRouter)
	api.Route("/session", route.SessionRouter)
	api.Route("/app", route.AppRouter)
	api.Route("/metrics", route.MetricRouter)

	// Frontend
	app.Use(
		"/*",
		func(ctx *fiber.Ctx) error {
			path := ctx.Path()

			// Redirect if url trails `/` (eg /login/ -> /login)
			if path != "/" && strings.HasSuffix(path, "/") {
				ctx.Redirect(strings.TrimSuffix(ctx.Request().URI().String(), "/"))
			} else {
				ctx.Next()
			}

			return nil
		},
		filesystem.New(filesystem.Config{
			Root: http.FS(frontend.FrontendFS),
			Browse: false,
			MaxAge: 86400,
			PathPrefix: "build",
			Next: func(ctx *fiber.Ctx) bool {
				// Auto add .html to extensionless paths (eg /login -> /login.html)
				path := ctx.Path()
				fileIndex := strings.LastIndex(path, "/")

				if path != "/" && !strings.Contains(path[fileIndex:], ".") {
					ctx.Path(path + ".html")
				}

				return false
			},
		}),
	)

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
