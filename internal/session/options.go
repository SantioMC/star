package session

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type Options struct {
	User            *database.User
	Account         *database.Account
	Application     *database.Application
	Scopes          []string
	Duration        time.Duration
	AuthenticatedAt time.Time
}
