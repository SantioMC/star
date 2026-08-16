package token

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

var ErrDurationTooLong = errors.New("access token duration may not be over 30 days")
var ErrNoActiveSigningKey = errors.New("there is no active signing key")

func Create(
	ctx context.Context,
	server *state.StarServer,
	session *database.Session,
	options *Options,
) (JWT, error) {
	if options.Duration > time.Hour*24*30 {
		return "", ErrDurationTooLong
	}

	fields := map[string]any{
		"sub": session.UserID,
		"sid": session.ID,
		"aud": session.ApplicationID,
	}

	if len(options.Scope) > 0 {
		fields["scope"] = strings.Join(options.Scope, " ")
	}

	if options.SuperUser {
		fields["permissions"] = []string{"*"}
	} else if options.Permissions != nil {
		fields["permissions"] = options.Permissions
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
