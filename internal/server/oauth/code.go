package oauth

import (
	"net/url"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/state"
)

func authorizationCodeGrant(ctx fiber.Ctx, form url.Values, application *database.Application) error {
	star := common.GetState(ctx)

	if application == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing client_id, clients must identify themselves")
	}

	code := form.Get("code")
	if code == "" {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing code")
	}

	authCode := star.Database.GetAuthorizationCodeByCode(ctx, code)
	if authCode == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired authorization code")
	}

	if authCode.ApplicationID != application.ID {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This authorization code was not issued to this client")
	}

	if authCode.RedeemedAt != nil {
		return revokeReplayedCode(ctx, star, authCode)
	}

	if authCode.ExpiresAt.Before(time.Now()) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired authorization code")
	}

	redirectURI := form.Get("redirect_uri")
	if redirectURI == "" || redirectURI != authCode.RedirectURI {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The redirect_uri does not match the authorization request")
	}

	if ok, err := verifyPKCE(ctx, application, authCode, form.Get("code_verifier")); !ok {
		return err
	}

	return redeemGrant(ctx, application, &issuedGrant{
		label:           "authorization code",
		userID:          authCode.UserID,
		impersonatorID:  authCode.ImpersonatorID,
		scopes:          authCode.Scopes,
		nonce:           authCode.Nonce,
		authenticatedAt: authCode.AuthenticatedAt,
		redeem: func(familyID string) (bool, error) {
			return star.Database.RedeemAuthorizationCode(ctx, authCode.ID, familyID)
		},
		replayed: func() error {
			return revokeReplayedCode(ctx, star, star.Database.GetAuthorizationCodeByID(ctx, authCode.ID))
		},
	})
}

func verifyPKCE(ctx fiber.Ctx, application *database.Application, authCode *database.AuthorizationCode, verifier string) (bool, error) {
	if authCode.CodeChallenge == nil {
		if verifier != "" {
			return false, writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "No code_challenge was included in the authorization request")
		}

		if application.Kind == database.ApplicationKindPublic {
			return false, writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Public clients must use PKCE")
		}

		return true, nil
	}

	if verifier == "" {
		return false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing code_verifier")
	}

	if !ValidCodeVerifierFormat(verifier) {
		return false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "The code_verifier must be 43-128 characters of [A-Za-z0-9-._~]")
	}

	if !VerifyCodeVerifier(verifier, *authCode.CodeChallenge) {
		return false, writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The code_verifier does not match the code_challenge")
	}

	return true, nil
}

func revokeReplayedCode(ctx fiber.Ctx, star *state.StarServer, authCode *database.AuthorizationCode) error {
	if authCode == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This authorization code has already been redeemed")
	}

	return revokeReplayedGrant(ctx, star, &grantReplay{
		kind:          database.SecurityAlertAuthorizationCodeUsed,
		label:         "authorization code",
		message:       "An authorization code was redeemed more than once, the session issued from it has been revoked",
		id:            authCode.ID,
		userID:        authCode.UserID,
		applicationID: authCode.ApplicationID,
		familyID:      authCode.FamilyID,
	})
}
