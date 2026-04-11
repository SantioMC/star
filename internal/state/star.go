package state

import (
	"github.com/tubblelive/star/internal/captcha"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/geolocation"
)

type StarServer struct {
	// Server data
	Database   *database.Database
	Config     *config.Config
	GeoService geolocation.Service
	Captcha    captcha.Verifier
	Email      email.EmailProvider

	// Cached data
	Settings    *database.Settings
	SigningKeys *[]database.SigningKey
}
