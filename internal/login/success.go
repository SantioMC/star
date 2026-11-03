package login

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

type SuccessfulLoginStage struct {
	Next        string            `json:"next"`
	AccessToken string            `json:"access_token"`
	Session     *database.Session `json:"-"`
}

func CreateSuccessfulStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
) (*SuccessfulLoginStage, error) {
	if err := star.Database.ClearLoginAttempts(ctx, user.ID); err != nil {
		return nil, err
	}

	scopes := []string{}
	for _, scope := range openid.BuiltinScopes {
		scopes = append(scopes, scope.Name)
	}

	session, err := session.CreateSession(
		ctx,
		star,
		&session.Options{
			User:     user,
			Account:  account,
			Duration: star.DefaultApp.Load().SessionDuration.Duration(),
			Scopes:   scopes,
		},
	)

	if err != nil {
		return nil, err
	}

	accessToken, err := IssueAccessToken(ctx, star, session, user)
	if err != nil {
		return nil, err
	}

	return &SuccessfulLoginStage{
		Next:        "complete",
		AccessToken: string(accessToken),
		Session:     session,
	}, nil
}

func IssueAccessToken(
	ctx fiber.Ctx,
	star *state.StarServer,
	session *database.Session,
	user *database.User,
) (token.JWT, error) {
	claims := openid.UserClaims(ctx, star.Database, user, session.Scopes)
	application := star.DefaultApp.Load()

	return token.Create(ctx, star, session, &token.Options{
		Application: application,
		Duration:    application.TokenDuration.Duration(),
		Scope:       session.Scopes,
		Claims:      claims,
	})
}
