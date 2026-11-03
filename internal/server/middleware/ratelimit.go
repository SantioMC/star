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
