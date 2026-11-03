package token

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

const tokenIDLength = 32

var ErrDurationTooLong = errors.New("access token duration may not be over 1 hour")
var ErrSessionRevoked = errors.New("the session the token is issued for has been revoked")
var ErrNoActiveSigningKey = errors.New("there is no active signing key")
var ErrMissingApplication = errors.New("access tokens must be issued to an application")
var ErrApplicationMismatch = errors.New("the application does not own the session the token is issued for")

var protectedClaims = []string{
	"iss",
	"sub",
	"aud",
	"azp",
	"client_id",
	"exp",
	"nbf",
	"iat",
	"naf",
	"jti",
	"sid",
	"auth_time",
	"nonce",
	"scope",
}

func Create(
	ctx context.Context,
	server *state.StarServer,
	session *database.Session,
	options *Options,
) (JWT, error) {
	if options.Duration > time.Hour {
		return "", ErrDurationTooLong
	}

	if options.Application == nil {
		return "", ErrMissingApplication
	}

	if options.Application.ID != session.ApplicationID {
		return "", ErrApplicationMismatch
	}

	audience := options.Audience
	if len(audience) == 0 {
		audience = []string{session.ApplicationID}
	}

	fields := make(map[string]any, len(options.Claims))
	for name, value := range options.Claims {
		if slices.Contains(protectedClaims, name) {
			slog.Warn("attempted to override protected claim", "key", name)
			continue
		}

		fields[name] = value
	}

	fields["sub"] = session.UserID
	fields["sid"] = session.FamilyID
	fields["aud"] = audience
	fields["jti"] = cryptography.CreateSecureString(tokenIDLength)
	fields["client_id"] = options.Application.ClientID
	fields["auth_time"] = session.AuthenticatedAt.Unix()

	if len(options.Scope) > 0 {
		fields["scope"] = strings.Join(options.Scope, " ")
	}

	issuedAt := time.Now()
	if server.Revocations.Revoked(issuedAt.Unix(), session.FamilyID) {
		return "", ErrSessionRevoked
	}

	if server.Revocations.Revoked(issuedAt.Unix(), session.UserID, database.GlobalRevocationSubject) {
		issuedAt = issuedAt.Add(time.Second)
	}

	token, err := createJWT(server, JWTOptions{
		Type:      TypeAccessToken,
		Algorithm: options.Application.SigningAlgorithm,
		Duration:  options.Duration,
		IssuedAt:  issuedAt,
		Fields:    fields,
	})

	if err != nil {
		return "", err
	}

	return token, nil
}
