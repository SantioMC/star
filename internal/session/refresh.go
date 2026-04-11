package session

import (
	"errors"
	"time"

	"github.com/tubblelive/star/internal/database"
)

var ErrSessionExpired = errors.New("this session has expired")

func RefreshSession(session *database.Session) (*database.Session, error) {
	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	return session, nil
}
