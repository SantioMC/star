package server

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/token"
)

func oidcRouter(router fiber.Router) {
	router.Get("/userinfo", middleware.RequireAuth, getUserInfo)
}

func getUserInfo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt := ctx.Locals(middleware.RequestAuthJWTKey).(*token.AccessTokenJWT)
	if jwt == nil {
		return errors.New("no jwt found in request, missing middleware?")
	}

	user := star.Database.GetUserByID(ctx, jwt.Subject)
	if user == nil {
		return errors.New("user not found from jwt")
	}

	email := star.Database.GetPrimaryEmail(ctx, user.ID)
	oauth := transformer.TransformOAuthUser(user, email)

	return ctx.JSON(oauth)
}
