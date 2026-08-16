package login

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

type SuccessfulLoginStage struct {
	Next         string `json:"next"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func CreateSuccessfulStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
) (*SuccessfulLoginStage, error) {
	permissions := []string{}
	for _, permission := range star.Database.GetUserPermissions(ctx, user.ID) {
		permissions = append(permissions, permission.Name)
	}

	session, err := session.CreateSession(
		ctx,
		star,
		&session.Options{
			User:     user,
			Account:  account,
			Duration: time.Hour * 24 * 7,
		},
	)

	if err != nil {
		return nil, err
	}

	accessToken, err := token.Create(ctx, star, session, &token.Options{
		Duration:    time.Hour,
		Permissions: permissions,
	})

	if err != nil {
		return nil, err
	}

	return &SuccessfulLoginStage{
		Next:         "complete",
		AccessToken:  string(accessToken),
		RefreshToken: session.Token,
	}, nil
}
