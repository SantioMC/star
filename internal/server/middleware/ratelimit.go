package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/tubblelive/star/internal/server/common"
)

func RateLimit(requests int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        requests,
		Expiration: time.Minute,
		KeyGenerator: func(ctx fiber.Ctx) string {
			return common.GetState(ctx).ClientIP(ctx)
		},
		LimitReached: func(ctx fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	})
}

func PollRateLimit(requests int, key func(fiber.Ctx) string, limited fiber.Handler) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          requests,
		Expiration:   time.Minute,
		KeyGenerator: key,
		LimitReached: limited,
		Next: func(ctx fiber.Ctx) bool {
			return key(ctx) == ""
		},
	})
}

func SkipWhen(skip func(fiber.Ctx) bool, handler fiber.Handler) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if skip(ctx) {
			return ctx.Next()
		}

		return handler(ctx)
	}
}

func UserFailureLimit(failures int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    failures,
		Expiration:             window,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(ctx fiber.Ctx) string {
			jwt, err := common.RequestJWT(ctx)
			if err != nil {
				return common.GetState(ctx).ClientIP(ctx)
			}

			return jwt.Subject
		},
		LimitReached: func(ctx fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	})
}
