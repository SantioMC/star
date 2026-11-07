package login

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

type RequireBirthdayStage struct {
	Next string `json:"next"`
	Flow string `json:"flow"`
}

func CreateRequireBirthdayStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
) (*RequireBirthdayStage, error) {
	token, err := createFlow(ctx, star, user, account, database.StageBirthday)
	if err != nil {
		return nil, err
	}

	return &RequireBirthdayStage{
		Next: "birthday",
		Flow: token,
	}, nil
}
