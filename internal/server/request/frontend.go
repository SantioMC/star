package request

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/frontend"
)

func ShowFrontend(ctx fiber.Ctx) error {
	return ctx.SendFile("build/index.html", fiber.SendFile{
		FS: frontend.FrontendFS,
	})
}
