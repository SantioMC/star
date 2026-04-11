package session

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type Options struct {
	User     *database.User
	Account  *database.Account
	Duration time.Duration
}
