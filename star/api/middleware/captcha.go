package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/state"
)

func RequireCaptcha(ctx *fiber.Ctx) error {
	captcha := state.StarState.Config.CaptchaVerifier
	if captcha == nil {
		return ctx.Next()
	}

	token := ctx.Get("X-Captcha-Token", "")
	if token == "" {
		return helper.NewHttpError(400, "Missing captcha token")
	}

	if len(token) > 2048 {
		return helper.NewHttpError(400, "Invalid captcha token")
	}

	valid := (*captcha).ValidateToken(token)
	if !valid {
		return helper.NewHttpError(400, "Invalid captcha token provided")
	}

	return ctx.Next()
}