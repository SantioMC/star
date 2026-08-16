package login

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/state"
)

type RequireMFAStage struct {
	Next    string          `json:"next"`
	Flow    string          `json:"flow"`
	Factors []mfa.MFAMethod `json:"factors"`
}

func CreateRequireMFAStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
	mfa []mfa.MFAMethod,
) (*RequireMFAStage, error) {
	id := database.CreateID(database.LoginFlowKey)

	token := database.CreateLoginFlowToken()
	if err := star.Database.CreateLoginFlow(ctx, &database.LoginFlow{
		ID:        id,
		UserID:    user.ID,
		Token:     token,
		Stage:     database.StageMFARequired,
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, err
	}

	return &RequireMFAStage{
		Next:    "mfa_required",
		Flow:    token,
		Factors: mfa,
	}, nil
}
