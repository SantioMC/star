package request

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/token"
)

// This method does not validate if the user exists on db, do not use for authentication endpoints
func GetJWT(ctx fiber.Ctx) (*token.AccessTokenJWT, error) {
	value := ctx.Locals(common.RequestAuthJWTKey)
	if value == nil {
		return nil, errors.New("missing jwt from request")
	}

	jwt := value.(*token.AccessTokenJWT)
	if jwt == nil {
		return nil, errors.New("value in request is not jwt")
	}

	return jwt, nil
}

// This method does not validate if the user exists on db, do not use for authentication endpoints
func GetUserID(ctx fiber.Ctx) (string, error) {
	// we can add api key support here down the road
	jwt, err := GetJWT(ctx)
	if err != nil {
		return "", err
	}

	return jwt.Subject, nil
}

func GetUser(ctx fiber.Ctx) (*database.User, error) {
	value := ctx.Locals(common.UserKey)
	if value != nil {
		data, ok := value.(*database.User)
		if ok {
			return data, nil
		}
	}

	star := common.GetState(ctx)
	userID, err := GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	user := star.Database.GetUserByID(ctx, userID)
	if user == nil {
		return nil, errors.New("user not found in database")
	}

	ctx.Locals(common.UserKey, user)
	return user, nil
}
