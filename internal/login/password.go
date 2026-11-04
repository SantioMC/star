package login

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

type RequirePasswordChangeStage struct {
	Next string `json:"next"`
	Flow string `json:"flow"`
}

func CreateRequirePasswordChangeStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
) (*RequirePasswordChangeStage, error) {
	token, err := createFlow(ctx, star, user, account, database.StagePasswordChange)
	if err != nil {
		return nil, err
	}

	return &RequirePasswordChangeStage{
		Next: "pass_change",
		Flow: token,
	}, nil
}
