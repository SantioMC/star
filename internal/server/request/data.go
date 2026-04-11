package request

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/token"
)

var userKey = "request.user"

func GetUser(ctx fiber.Ctx) (*database.User, error) {
	data, ok := ctx.Locals(userKey).(*database.User)
	if ok {
		return data, nil
	}

	star := common.GetState(ctx)
	jwt := ctx.Locals(middleware.RequestAuthJWTKey).(*token.AccessTokenJWT)
	if jwt == nil {
		return nil, errors.New("missing jwt from request")
	}

	user := star.Database.GetUserByID(ctx, jwt.Subject)
	if user == nil {
		return nil, errors.New("user not found in database")
	}

	ctx.Locals(userKey, user)
	return user, nil
}
