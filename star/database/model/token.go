package model

import (
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
	"gorm.io/datatypes"
)

type RefreshToken struct {
	BaseModel

	UserID        string
	User          *User
	ApplicationID string
	Application   *Application

	Token       string
	Fingerprint *string
	Claims      datatypes.JSONType[[]oidc.Claim]
	Duration    time.Duration

	InvalidatedAt *int64
}

func NewRefreshToken(duration time.Duration, claims []oidc.Claim) RefreshToken {
	token := crypto.CreateRandString(48)

	return RefreshToken{
		BaseModel: BaseModel{
			ID: CreateTypedId("rt", 12),
		},
		Token:    token,
		Duration: duration,
		Claims:   datatypes.NewJSONType(claims),
	}
}

func (t *RefreshToken) IsExpired() bool {
	expiry := t.CreatedAt + int64(t.Duration.Seconds())
	return time.Now().After(time.Unix(expiry, 0))
}
