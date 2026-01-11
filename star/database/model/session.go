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

	Fingerprint *string `encrypt:"true"`
	UserAgent   string  `encrypt:"true"`
	Address     string  `encrypt:"true"`

	Claims   datatypes.JSONType[[]oidc.Claim]
	Duration time.Duration

	InvalidatedAt *int64
	RotatedAt     *int64
}

func NewSession(duration time.Duration, claims []oidc.Claim) Session {
	return Session{
		BaseModel: BaseModel{
			ID: CreateTypedId("ses", 12),
		},
		Token:    NewSessionToken(),
		Duration: duration,
		Claims:   datatypes.NewJSONType(claims),
	}
}

func CopySession(previous Session) Session {
	return Session{
		BaseModel: BaseModel{
			ID: CreateTypedId("ses", 12),
		},

		Token: NewSessionToken(),

		UserID:        previous.UserID,
		User:          previous.User,
		ApplicationID: previous.ApplicationID,
		Application:   previous.Application,

		Duration:    previous.Duration,
		Claims:      previous.Claims,
		Fingerprint: previous.Fingerprint,
		UserAgent:   previous.UserAgent,
		Address:     previous.Address,
	}
}

func NewSessionToken() string {
	return crypto.CreateRandString(48)
}

func (s *Session) ExpiresAt() int64 {
	return s.CreatedAt + int64(s.Duration.Seconds())
}

func (s *Session) IsExpired() bool {
	return time.Now().Unix() > s.ExpiresAt() || s.InvalidatedAt != nil
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
