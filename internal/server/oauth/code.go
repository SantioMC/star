package oauth

import (
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
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

	if authCode.ExpiresAt.Before(time.Now()) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired authorization code")
	}

	if authCode.RedeemedAt != nil {
		return revokeReplayedCode(ctx, star, authCode)
	}

	redirectUri := form.Get("redirect_uri")
	if redirectUri == "" || redirectUri != authCode.RedirectURI {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The redirect_uri does not match the authorization request")
	}

	if ok, err := verifyPKCE(ctx, application, authCode, form.Get("code_verifier")); !ok {
		return err
	}

	user := star.Database.GetUserByID(ctx, authCode.UserID)
	if user == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this authorization code was issued for no longer exists")
	}

	newSession, err := session.CreateSession(ctx, star, &session.Options{
		User:            user,
		Application:     application,
		Scopes:          authCode.Scopes,
		Duration:        application.SessionDuration.Duration(),
		AuthenticatedAt: authCode.AuthenticatedAt,
	})

	if err != nil {
		return err
	}

	redeemed, err := star.Database.RedeemAuthorizationCode(ctx, authCode.ID, newSession.ID)
	if err != nil {
		return err
	}

	if !redeemed {
		if err := star.Database.DeleteSessionByID(ctx, newSession.ID); err != nil {
			slog.Error("failed to delete session after losing authorization code redemption race", "err", err, "session", newSession.ID)
		}

		return revokeReplayedCode(ctx, star, star.Database.GetAuthorizationCodeByID(ctx, authCode.ID))
	}

	accessToken, err := token.Create(ctx, star, newSession, &token.Options{
		Duration: application.TokenDuration.Duration(),
		Scope:    authCode.Scopes,
	})

	if err != nil {
		return err
	}

	response := tokenResponse{
		AccessToken: string(accessToken),
		TokenType:   "Bearer",
		ExpiresIn:   int64(application.TokenDuration.Duration().Seconds()),
		Scope:       strings.Join(authCode.Scopes, " "),
	}

	if slices.Contains(authCode.Scopes, openid.ScopeOfflineAccess) {
		response.RefreshToken = newSession.Token
	}

	if slices.Contains(authCode.Scopes, openid.ScopeOpenID) {
		claims := openid.UserClaims(ctx, star.Database, user, authCode.Scopes)

		idToken, err := token.CreateIDToken(star, newSession, application, claims, authCode.Nonce)
		if err != nil {
			return err
		}

		response.IDToken = string(idToken)
	}

	return writeTokenResponse(ctx, response)
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

	slog.Warn("authorization code replay detected, revoking the previously issued session", "code", authCode.ID, "application", authCode.ApplicationID)

	if authCode.SessionID != nil {
		if err := star.Database.DeleteSessionByID(ctx, *authCode.SessionID); err != nil {
			slog.Error("failed to revoke session for replayed authorization code", "err", err, "session", *authCode.SessionID)
		}
	}

	return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This authorization code has already been redeemed")
}
