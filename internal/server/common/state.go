package common

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/state"
)

var StarKey = "server.state"
var UserKey = "request.user"
var RequestAuthJWTKey = "request.auth.jwt"
var RequestWeightKey = "request.weight"

func GetState(ctx fiber.Ctx) *state.StarServer {
	data, ok := ctx.Locals(StarKey).(*state.StarServer)
	if !ok {
		panic("failed to get state from context")
	}

	return data
}
