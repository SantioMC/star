package passkey

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/security"
	"github.com/tubblelive/star/internal/state"
)

func BeginLogin(ctx context.Context, star *state.StarServer) (string, *protocol.PublicKeyCredentialRequestOptions, error) {
	assertion, session, err := star.WebAuthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)

	if err != nil {
		return "", nil, err
	}

	token, err := storeChallenge(ctx, star, session, &database.PasskeyChallenge{
		Kind: database.PasskeyChallengeLogin,
	})

	return token, &assertion.Response, err
}

func BeginMFA(ctx context.Context, star *state.StarServer, flow *database.LoginFlow) (string, *protocol.PublicKeyCredentialRequestOptions, error) {
	owner, err := loadUser(ctx, star, flow.UserID)
	if err != nil {
		return "", nil, err
	}

	assertion, session, err := star.WebAuthn.BeginLogin(
		owner,
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)

	if err != nil {
		return "", nil, err
	}

	token, err := storeChallenge(ctx, star, session, &database.PasskeyChallenge{
		Kind:        database.PasskeyChallengeMFA,
		UserID:      &flow.UserID,
		LoginFlowID: &flow.ID,
	})

	return token, &assertion.Response, err
}

func FinishLogin(ctx context.Context, star *state.StarServer, token string, response []byte) (string, error) {
	challenge, parsed, err := consumeChallenge(ctx, star, token, database.PasskeyChallengeLogin, response)
	if err != nil {
		return "", err
	}

	account, credential, err := star.WebAuthn.ValidatePasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			return loadUser(ctx, star, string(userHandle))
		},
		challenge.Session.Data,
		parsed,
	)

	if err != nil {
		slog.Warn("passkey login failed", "reason", "verification failed", "err", err)
		return "", ErrVerificationFailed
	}

	owner, ok := account.(*user)
	if !ok {
		return "", errors.New("passkey login returned an unexpected user type")
	}

	if err := recordUsage(ctx, star, owner, credential); err != nil {
		return "", err
	}

	return owner.user.ID, nil
}

func FinishMFA(ctx context.Context, star *state.StarServer, token string, flow *database.LoginFlow, response []byte) error {
	challenge, parsed, err := consumeChallenge(ctx, star, token, database.PasskeyChallengeMFA, response)
	if err != nil {
		return err
	} else if challenge.LoginFlowID == nil || *challenge.LoginFlowID != flow.ID {
		return ErrInvalidChallenge
	}

	owner, err := loadUser(ctx, star, flow.UserID)
	if err != nil {
		return err
	}

	credential, err := star.WebAuthn.ValidateLogin(owner, challenge.Session.Data, parsed)
	if err != nil {
		slog.Warn("passkey mfa failed", "reason", "verification failed", "user", flow.UserID, "err", err)
		return ErrVerificationFailed
	}

	return recordUsage(ctx, star, owner, credential)
}

func storeChallenge(
	ctx context.Context,
	star *state.StarServer,
	session *webauthn.SessionData,
	challenge *database.PasskeyChallenge,
) (string, error) {
	challenge.ID = database.CreateID(database.PasskeyChallengeKey)
	challenge.Token = database.CreatePasskeyChallengeToken()
	challenge.Session = models.JSONData[webauthn.SessionData]{Data: *session}
	challenge.CreatedAt = time.Now()

	if err := star.Database.CreatePasskeyChallenge(ctx, challenge); err != nil {
		return "", err
	}

	return challenge.Token, nil
}

func consumeChallenge(
	ctx context.Context,
	star *state.StarServer,
	token string,
	kind database.PasskeyChallengeKind,
	response []byte,
) (*database.PasskeyChallenge, *protocol.ParsedCredentialAssertionData, error) {
	challenge, err := star.Database.ConsumePasskeyChallenge(ctx, token, kind)
	if err != nil {
		return nil, nil, err
	} else if challenge == nil {
		return nil, nil, ErrInvalidChallenge
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		slog.Warn("passkey login failed", "reason", "malformed response", "err", err)
		return nil, nil, ErrVerificationFailed
	}

	exists, err := star.Database.PasskeyExists(ctx, parsed.RawID)
	if err != nil {
		return nil, nil, err
	} else if !exists {
		slog.Warn("passkey login failed", "reason", "unknown credential")
		return nil, nil, ErrUnknownPasskey
	}

	return challenge, parsed, nil
}

func loadUser(ctx context.Context, star *state.StarServer, userID string) (*user, error) {
	account := star.Database.GetUserByID(ctx, userID)
	if account == nil {
		return nil, errors.New("user does not exist")
	}

	return newUser(ctx, star, account)
}

func recordUsage(ctx context.Context, star *state.StarServer, owner *user, credential *webauthn.Credential) error {
	passkey := owner.findPasskey(credential.ID)
	if passkey == nil {
		return errors.New("validated passkey credential was not found for user")
	}

	if credential.Authenticator.CloneWarning {
		security.Report(star, ctx, &database.SecurityAlert{
			Kind:     database.SecurityAlertPasskeyCloned,
			Severity: database.SecurityAlertSeverityHigh,
			Message:  "Passkey signature counter did not increase, the authenticator may have been cloned",
			UserID:   &owner.user.ID,
			Data:     map[string]any{"passkey": passkey.ID},
		})
	}

	now := time.Now()
	passkey.LastUsedAt = &now
	passkey.Credential = models.JSONData[webauthn.Credential]{Data: *credential}

	return star.Database.UpdatePasskeyUsage(ctx, passkey)
}
