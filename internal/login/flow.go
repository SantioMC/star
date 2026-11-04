package login

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func createFlow(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
	stage database.LoginFlowStage,
) (string, error) {
	var accountID *string
	if account != nil {
		accountID = &account.ID
	}

	token := database.CreateLoginFlowToken()
	err := star.Database.CreateLoginFlow(ctx, &database.LoginFlow{
		ID:        database.CreateID(database.LoginFlowKey),
		UserID:    user.ID,
		AccountID: accountID,
		Token:     token,
		Stage:     stage,
		CreatedAt: time.Now(),
	})

	return token, err
}
