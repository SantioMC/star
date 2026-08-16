package oauth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
)

type PublicApplication struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Implicit bool   `json:"implicit"`
}

func Application(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByClientID(ctx, ctx.Params("client_id"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	return ctx.JSON(PublicApplication{
		ID:       application.ID,
		Name:     application.Name,
		Implicit: application.Internal,
	})
}
