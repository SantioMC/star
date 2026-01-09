package model

import (
	"strings"
	"time"
)

const DefaultAccessTokenDuration = time.Minute * 5
const DefaultRefreshTokenDuration = time.Hour * 24 * 7

type Application struct {
	BaseModel

	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Url      string `json:"url"`
	Callback string `json:"callback"`
	Internal bool   `json:"internal"`

	RotateRefreshToken   bool  `json:"rotate_refresh_token"`
	TokenDuration        int64 `json:"token_duration"`
	RefreshTokenDuration int64 `json:"refresh_token_duration"`

	Providers []Provider `json:"providers"`
}

func NewApplication(name string) Application {
	slug := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	return Application{
		BaseModel: BaseModel{
			ID: CreateTypedId("app", 12),
		},

		Name:     name,
		Slug:     slug,
		Url:      "",
		Callback: "",
		Internal: false,

		// Base Settings
		RotateRefreshToken:   true,
		TokenDuration:        int64(DefaultAccessTokenDuration.Seconds()),
		RefreshTokenDuration: int64(DefaultRefreshTokenDuration.Seconds()),

		Providers: []Provider{},
	}
}
