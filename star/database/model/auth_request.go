package model

import (
	"time"

	"github.com/tubblelive/star/star/crypto"
	"gorm.io/datatypes"
)

type AuthRequest struct {
	BaseModel

	ApplicationID string       `json:"application_id"`
	Application   *Application `json:"-"`

	UserID string `json:"user_id"`
	User   *User  `json:"-"`

	Code        string  `json:"code"`
	RedirectURI *string `json:"callback_uri"`
	State       *string `json:"state"`
	Fingerprint *string `json:"fingerprint"`
	Scopes      datatypes.JSONType[[]string]

	ExpiresAt   int64  `json:"expires_at"`
	CompletedAt *int64 `json:"completed_at"`
}

func (request *AuthRequest) IsExpired() bool {
	return request.ExpiresAt < time.Now().Unix()
}

func NewAuthRequest(
	application *Application,
	user *User,
	scopes []string,
	callbackUri *string,
	fingerprint *string,
	state *string,
) AuthRequest {
	expiresAt := time.Now().Add(time.Minute * 10)

	return AuthRequest{
		BaseModel: BaseModel{
			ID: CreateTypedId("auth", 12),
		},

		Application:   application,
		ApplicationID: application.ID,

		User:   user,
		UserID: user.ID,

		Code:        crypto.CreateRandString(24),
		State:       state,
		Scopes:      datatypes.NewJSONType(scopes),
		RedirectURI: callbackUri,

		ExpiresAt: expiresAt.Unix(),
	}
}
