package model

import (
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
	"gorm.io/datatypes"
)

type Session struct {
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

func NewSession(duration time.Duration, claims []oidc.Claim) Session {
	token := crypto.CreateRandString(48)

	return Session{
		BaseModel: BaseModel{
			ID: CreateTypedId("rt", 12),
		},
		Token:    token,
		Duration: duration,
		Claims:   datatypes.NewJSONType(claims),
	}
}

func (t *Session) IsExpired() bool {
	expiry := t.CreatedAt + int64(t.Duration.Seconds())
	return time.Now().After(time.Unix(expiry, 0))
}
