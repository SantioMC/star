package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/tubblelive/star/frontend"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/state"
)

func StartServer(star *state.StarServer) error {
	app := fiber.New(fiber.Config{
		AppName:         "Star Authorization Server",
		ServerHeader:    "Star",
		BodyLimit:       1024 * 1024 * 10,
		ProxyHeader:     star.Config.ProxyHeader,
		TrustProxy:      star.Config.TrustProxy,
		StructValidator: request.NewStructValidator(),
		ReadTimeout:     time.Second * 10,
		WriteTimeout:    time.Second * 10,
		IdleTimeout:     time.Second * 15,
	})

	app.Hooks().OnPreStartupMessage(func(data *fiber.PreStartupMessageData) error {
		data.PreventDefault = true
		return nil
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: star.Config.Origins,
		AllowHeaders: []string{
			fiber.HeaderAuthorization,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderOrigin,
		},
		AllowCredentials: true,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodOptions,
			fiber.MethodPatch,
		},
	}))

	app.Use(func(ctx fiber.Ctx) error {
		ctx.Locals(common.StarKey, star)
		return ctx.Next()
	})

	app.Use(func(ctx fiber.Ctx) error {
		err := ctx.Next()

		uri := string(ctx.Request().URI().RequestURI())
		request := ctx.Method() + " " + uri

		if err != nil {
			switch data := err.(type) {
			case *fiber.Error:
				slog.Error(strconv.Itoa(data.Code)+" "+request, "err", err)
				ctx.Status(data.Code).JSON(fiber.Map{
					"error":   true,
					"status":  data.Code,
					"message": data.Message,
				})
			case validator.ValidationErrors:
				slog.Error(strconv.Itoa(400)+" "+request, "err", err)
				errors := []string{}

				for _, err := range data {
					errors = append(
						errors,
						fmt.Sprintf("Field '%s' must be %s", err.Field(), err.Tag()),
					)
				}

				ctx.Status(400).JSON(fiber.Map{
					"error":   true,
					"status":  400,
					"message": "Invalid request body provided",
					"errors":  errors,
				})
			default:
				params := map[string]any{}
				ctx.Bind().URI(&params)

				slog.Error("500 "+request, "params", params, "err", err)
				ctx.Status(500).JSON(fiber.Map{
					"error":   true,
					"status":  500,
					"message": data.Error(),
				})
			}
		} else if strings.HasPrefix(uri, "/api") {
			slog.Info(strconv.Itoa(ctx.Response().StatusCode()) + " " + request)
		}

		return nil
	})

	app.Use(compress.New())
	app.Use(recover.New())

	global := app.Group("/", cors.New(cors.Config{
		AllowCredentials: false,
	}))

	global.Route("/healthz", healthRouter)
	global.Route("/.well-known", wellKnownRouter)
	global.Route("/oauth", oauthRouter)
	global.Route("/oidc", oidcRouter)

	// API
	api := app.Group("/api", limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(ctx fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	api.Route("/email", emailRouter)
	api.Route("/user", userRouter)
	api.Use(middleware.RequireAuth).Route("/app", appRouter)
	api.Use(middleware.RequireAuth).Route("/session", sessionRouter)
	api.Use(middleware.RequireAuth).Route("/mfa", mfaRouter)
	api.Use(middleware.RequireAuth).Route("/role", roleRouter)
	api.Use(middleware.RequireAuth).Route("/provider", providerRouter)

	api.All("*", func(ctx fiber.Ctx) error {
		// We don't want API routes to fallback to the frontend and send a 200
		return fiber.ErrNotFound
	})

	// Frontend
	siteKey := ""

	if star.Config.Captcha.IsEnabled() {
		siteKey = star.Captcha.GetSiteKey()
	}

	environment, err := json.Marshal(map[string]any{
		"PUBLIC_API_URL":  star.Config.ServerURL,
		"PUBLIC_SITE_KEY": siteKey,
	})

	if err != nil {
		panic("failed to marshal frontend environment")
	}

	app.Get("/_app/env.js", func(ctx fiber.Ctx) error {
		// Cloudflare cached this and made my life miserable so cache is now 5mins
		ctx.Set(fiber.HeaderCacheControl, "public, max-age=300")
		ctx.Set(fiber.HeaderContentType, "application/javascript")
		return ctx.SendString("export const env=" + string(environment))
	})

	build, err := fs.Sub(frontend.FrontendFS, "build")
	if err != nil {
		return fmt.Errorf("failed to get build directory: %w", err)
	}

	app.Use(static.New("/_app/*", static.Config{
		FS:            build,
		CacheDuration: time.Hour * 24 * 7,
	}))

	app.Get("/*", func(ctx fiber.Ctx) error {
		return request.ShowFrontend(ctx)
	})

	port := fmt.Sprintf("[::]:%d", star.Config.Port)
	if err := app.Listen(port); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
