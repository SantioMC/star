package model

import (
	"time"

	"github.com/medama-io/go-useragent"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
	"gorm.io/datatypes"
)

var parser = useragent.NewParser()

type Session struct {
	BaseModel

	UserID        string
	User          *User
	ApplicationID string
	Application   *Application

	Token string

	Fingerprint *string
	UserAgent   string

	Claims   datatypes.JSONType[[]oidc.Claim]
	Duration time.Duration

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

func (s *Session) IsExpired() bool {
	expiry := s.CreatedAt + int64(s.Duration.Seconds())
	return time.Now().After(time.Unix(expiry, 0))
}

func (s *Session) GetOS() string {
	if s.UserAgent == "" {
		return ""
	}

	agent := parser.Parse(s.UserAgent)
	return agent.OS().String()
}

func (s *Session) GetDevice() string {
	if s.UserAgent == "" {
		return ""
	}

	agent := parser.Parse(s.UserAgent)
	return agent.Device().String()
}