package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

var HeavyRateLimiter = limiter.Config{
	Max:        10,
	Expiration: 1 * time.Minute,
	LimitReached: func(ctx fiber.Ctx) error {
		return fiber.ErrTooManyRequests
	},
}
