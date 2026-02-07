package service

import (
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

func GetAuthRequestByCode(code string) (*model.AuthRequest, error) {
	return state.Database.OAuthStore.GetAuthRequestByCode(code)
}

func CreateAuthRequest(
	app *model.Application,
	user *model.User,
	scopes []string,
	callbackUri string,
	nonce *string,
) (model.AuthRequest, error) {
	if len(scopes) == 0 {
		scopes = oidc.DefaultScopes
	}

	return state.Database.OAuthStore.CreateAuthRequest(app, user, scopes, callbackUri, nonce)
}

func MarkAuthRequestCompleted(code string) error {
	return state.Database.OAuthStore.MarkCompleted(code)
}
