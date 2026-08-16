package token

import (
	"maps"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func CreateIDToken(
	server *state.StarServer,
	session *database.Session,
	application *database.Application,
	claims map[string]any,
	nonce *string,
) (JWT, error) {
	fields := maps.Clone(claims)
	if fields == nil {
		fields = map[string]any{}
	}

	fields["sub"] = session.UserID
	fields["sid"] = session.ID
	fields["aud"] = application.ClientID
	fields["azp"] = application.ClientID
	fields["auth_time"] = session.AuthenticatedAt.Unix()

	if nonce != nil {
		fields["nonce"] = *nonce
	}

	return createJWT(server, JWTBody{
		Duration: application.TokenDuration.Duration(),
		Fields:   fields,
	})
}
