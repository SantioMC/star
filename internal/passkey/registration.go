package passkey

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/state"
)

var ErrInvalidChallenge = fiber.NewError(400, "Passkey request has expired, please try again")
var ErrVerificationFailed = fiber.NewError(401, "Unable to verify your passkey, please try again")
var ErrUnknownPasskey = fiber.NewError(404, "This passkey is no longer registered to an account")

func BeginRegistration(ctx context.Context, star *state.StarServer, account *database.User) (string, *protocol.PublicKeyCredentialCreationOptions, error) {
	owner, err := newUser(ctx, star, account)
	if err != nil {
		return "", nil, err
	}

	creation, session, err := star.WebAuthn.BeginRegistration(
		owner,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: new(true),
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}),
		webauthn.WithExclusions(webauthn.Credentials(owner.WebAuthnCredentials()).CredentialDescriptors()),
	)

	if err != nil {
		return "", nil, err
	}

	token, err := storeChallenge(ctx, star, session, &database.PasskeyChallenge{
		Kind:   database.PasskeyChallengeRegister,
		UserID: &account.ID,
	})

	return token, &creation.Response, err
}

func FinishRegistration(
	ctx context.Context,
	star *state.StarServer,
	account *database.User,
	token string,
	name string,
	response []byte,
) (*database.Passkey, error) {
	challenge, err := star.Database.ConsumePasskeyChallenge(ctx, token, database.PasskeyChallengeRegister)
	if err != nil {
		return nil, err
	} else if challenge == nil || challenge.UserID == nil || *challenge.UserID != account.ID {
		return nil, ErrInvalidChallenge
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(response)
	if err != nil {
		slog.Warn("passkey registration failed", "reason", "malformed response", "user", account.ID, "err", err)
		return nil, ErrVerificationFailed
	}

	owner, err := newUser(ctx, star, account)
	if err != nil {
		return nil, err
	}

	credential, err := star.WebAuthn.CreateCredential(owner, challenge.Session.Data, parsed)
	if err != nil {
		slog.Warn("passkey registration failed", "reason", "verification failed", "user", account.ID, "err", err)
		return nil, ErrVerificationFailed
	}

	passkey := &database.Passkey{
		ID:           database.CreateID(database.PasskeyKey),
		UserID:       account.ID,
		Name:         name,
		CredentialID: credential.ID,
		Credential:   models.JSONData[webauthn.Credential]{Data: *credential},
		CreatedAt:    time.Now(),
	}

	if err := star.Database.CreatePasskey(ctx, passkey); err != nil {
		return nil, err
	}

	return passkey, nil
}
