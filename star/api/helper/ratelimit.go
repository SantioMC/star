package helper

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

var HeavyLimiter = limiter.Config{
	Max:        10,
	Expiration: 1 * time.Minute,
	LimitReached: func(c *fiber.Ctx) error {
		return fiber.ErrTooManyRequests
	},
}