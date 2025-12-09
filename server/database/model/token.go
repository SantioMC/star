package model

import (
	"time"

	"github.com/tubblelive/star/server/crypto"
)

type RefreshToken struct {
	BaseModel
	UserID        string
	User          *User
	ApplicationID string
	Application   *Application
	Token         string
	Fingerprint   *string
	Duration      time.Duration
	InvalidatedAt *int64
}

func NewRefreshToken(duration time.Duration) RefreshToken {
	token := crypto.CreateRandString(24)

	return RefreshToken{
		BaseModel: BaseModel{
			ID: CreateTypedId("rt", 12),
		},
		Token:    token,
		Duration: duration,
	}
}

func (t *RefreshToken) IsExpired() bool {
	expiry := t.CreatedAt + int64(t.Duration.Seconds())
	return time.Now().After(time.Unix(expiry, 0))
}
