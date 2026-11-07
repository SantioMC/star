package types

import (
	"time"

	"github.com/tubblelive/star/internal/server/oauth"
)

type VerifyDeviceRequest struct {
	UserCode string `json:"user_code" validate:"required,max=32"`
}

type VerifyDeviceResponse struct {
	UserCode    string                  `json:"user_code"`
	Device      string                  `json:"device"`
	Geolocation *string                 `json:"geolocation"`
	ExpiresAt   time.Time               `json:"expires_at"`
	Application oauth.PublicApplication `json:"application"`
}

type DecideDeviceRequest struct {
	UserCode string `json:"user_code" validate:"required,max=32"`
	Approve  bool   `json:"approve"`
}
