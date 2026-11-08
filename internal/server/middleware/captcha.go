package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
)

func RequireCaptcha(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	if !star.CaptchaConfig().IsEnabled() {
		return ctx.Next()
	}

	token := ctx.Get("X-Captcha-Token", "")
	if token == "" {
		return fiber.NewError(400, "Missing captcha token")
	}

	if len(token) > 2048 {
		return fiber.NewError(400, "Invalid captcha token")
	}

	valid := star.Captcha().ValidateToken(token)
	if !valid {
		return fiber.NewError(400, "Invalid captcha token provided")
	}

	return ctx.Next()
}
