package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
)

func oidcRouter(router fiber.Router) {
	router.Get("/userinfo", middleware.RequireAuth, getUserInfo)
}

func getUserInfo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetPrimaryEmail(ctx, user.ID)
	oauth := transformer.TransformOAuthUser(user, email)

	return ctx.JSON(oauth)
}
