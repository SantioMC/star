package token

import (
	"context"
	"errors"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

var ErrDurationTooLong = errors.New("access token duration may not be over a day")
var ErrNoActiveSigningKey = errors.New("there is no active signing key")

func Create(
	ctx context.Context,
	server *state.StarServer,
	session *database.Session,
	options *Options,
) (JWT, error) {
	if options.Duration > time.Hour*24 {
		return "", ErrDurationTooLong
	}

	fields := map[string]any{
		"sub": session.UserID,
		"sid": session.ID,
	}

	if options.Permissions != nil {
		fields["permissions"] = options.Permissions
	}

	if options.SuperUser {
		fields["permissions"] = append(fields["permissions"].([]string), "*")
	}

	token, err := createJWT(server, JWTBody{
		Duration: options.Duration,
		Fields:   fields,
	})

	if err != nil {
		return "", err
	}

	return token, nil
}
