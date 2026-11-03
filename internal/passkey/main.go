package passkey

import (
	"context"
	"net/url"
	"slices"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func NewRelyingParty(config *config.Config) (*webauthn.WebAuthn, error) {
	server, err := url.Parse(config.ServerURL)
	if err != nil {
		return nil, err
	}

	origins := []string{server.Scheme + "://" + server.Host}
	for _, origin := range config.Origins {
		if origin != "*" && !slices.Contains(origins, origin) {
			origins = append(origins, origin)
		}
	}

	return webauthn.New(&webauthn.Config{
		RPID:          server.Hostname(),
		RPDisplayName: "Star",
		RPOrigins:     origins,
	})
}

type user struct {
	user     *database.User
	passkeys []database.Passkey
}

func newUser(ctx context.Context, star *state.StarServer, account *database.User) (*user, error) {
	passkeys, err := star.Database.GetUserPasskeys(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	return &user{user: account, passkeys: passkeys}, nil
}

func UserHandle(account *database.User) []byte {
	return []byte(account.ID)
}

func (u *user) WebAuthnID() []byte {
	return UserHandle(u.user)
}

func (u *user) WebAuthnName() string {
	return u.user.Name
}

func (u *user) WebAuthnDisplayName() string {
	return u.user.Name
}

func (u *user) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, 0, len(u.passkeys))
	for _, passkey := range u.passkeys {
		credentials = append(credentials, passkey.Credential.Data)
	}

	return credentials
}

func (u *user) findPasskey(credentialID []byte) *database.Passkey {
	for idx := range u.passkeys {
		if slices.Equal(u.passkeys[idx].CredentialID, credentialID) {
			return &u.passkeys[idx]
		}
	}

	return nil
}
