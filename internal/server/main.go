package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/tubblelive/star/frontend"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/state"
)

func CreateServer(star *state.StarServer) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		AppName:      "Star Authorization Server",
		ServerHeader: "Star",
		BodyLimit:    1024 * 1024 * 10,
		ProxyHeader:  star.Config.Proxy.Header,
		TrustProxy:   star.Config.Proxy.Enabled,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: star.Config.Proxy.Trusted,
		},
		EnableIPValidation: true,
		StructValidator:    request.NewStructValidator(),
		ReadTimeout:        time.Second * 10,
		WriteTimeout:       time.Second * 10,
		IdleTimeout:        time.Second * 15,
	})

	app.Hooks().OnPreStartupMessage(func(data *fiber.PreStartupMessageData) error {
		data.PreventDefault = true
		return nil
	})

	public := []string{"/healthz", "/.well-known", "/oauth", "/oidc"}

	app.Use(cors.New(cors.Config{
		Next: func(ctx fiber.Ctx) bool {
			return slices.ContainsFunc(public, func(prefix string) bool {
				return ctx.Path() == prefix || strings.HasPrefix(ctx.Path(), prefix+"/")
			})
		},
		AllowOrigins: star.Config.Origins,
		AllowHeaders: []string{
			fiber.HeaderAuthorization,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderOrigin,
			"x-captcha-token",
		},
		ExposeHeaders: []string{
			fiber.HeaderWWWAuthenticate,
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
					"message": "An internal error has occurred, please try again later",
				})
			}
		} else if strings.HasPrefix(uri, "/api") {
			slog.Info(strconv.Itoa(ctx.Response().StatusCode()) + " " + request)
		}

		return nil
	})

	app.Use(compress.New())
	app.Use(recover.New())

	app.Use(public, cors.New(cors.Config{
		AllowCredentials: false,
	}))

	app.Route("/healthz", healthRouter)
	app.Route("/.well-known", wellKnownRouter)
	app.Route("/oauth", oauthRouter)
	app.Route("/oidc", oidcRouter)

	// API
	api := app.Group("/api", middleware.RateLimit(100))

	api.Route("/email", emailRouter)
	api.Route("/user", userRouter)
	api.Route("/app", appRouter)
	api.Route("/session", sessionRouter)
	api.Route("/grant", grantRouter)
	api.Route("/mfa", mfaRouter)
	api.Route("/role", roleRouter)
	api.Route("/provider", providerRouter)
	api.Route("/settings", settingsRouter)
	api.Route("/key", keyRouter)
	api.Route("/revocation", revocationRouter)
	api.Route("/metrics", metricsRouter)
	api.Route("/alert", alertRouter)

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
		return nil, fmt.Errorf("failed to marshal frontend environment: %w", err)
	}

	app.Get("/_app/env.js", func(ctx fiber.Ctx) error {
		// Cloudflare cached this and made my life miserable so cache is now 5mins
		ctx.Set(fiber.HeaderCacheControl, "public, max-age=300")
		ctx.Set(fiber.HeaderContentType, "application/javascript")
		return ctx.SendString("export const env=" + string(environment))
	})

	build, err := fs.Sub(frontend.FrontendFS, "build")
	if err != nil {
		return nil, fmt.Errorf("failed to get build directory: %w", err)
	}

	app.Use(static.New("/_app/*", static.Config{
		FS:            build,
		CacheDuration: time.Hour * 24 * 7,
	}))

	app.Get("/*", func(ctx fiber.Ctx) error {
		return request.ShowFrontend(ctx)
	})

	return app, nil
}
