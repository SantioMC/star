package common

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/token"
)

func RequestJWT(ctx fiber.Ctx) (*token.AccessTokenJWT, error) {
	value := ctx.Locals(RequestAuthJWTKey)
	if value == nil {
		return nil, errors.New("no jwt found in request, there is likely no authentication middleware before this one")
	}

	jwt, ok := value.(*token.AccessTokenJWT)
	if !ok || jwt == nil {
		return nil, errors.New("the request auth jwt key does not hold a jwt")
	}

	return jwt, nil
}
