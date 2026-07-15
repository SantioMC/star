package login

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/state"
)

type RequireMFAStage struct {
	Next        string          `json:"next"`
	LoginFlowID string          `json:"flow_id"`
	Factors     []mfa.MFAMethod `json:"factors"`
}

func CreateRequireMFAStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
	mfa []mfa.MFAMethod,
) (*RequireMFAStage, error) {
	id := database.CreateID(database.LoginFlowKey)

	if err := star.Database.CreateLoginFlow(ctx, &database.LoginFlow{
		ID:        id,
		UserID:    user.ID,
		Stage:     database.StageMFARequired,
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, err
	}

	return &RequireMFAStage{
		Next:        "mfa_required",
		LoginFlowID: id,
		Factors:     mfa,
	}, nil
}
