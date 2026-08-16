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
	"github.com/tubblelive/star/internal/token"
)

func refreshTokenGrant(ctx fiber.Ctx, form url.Values, application *database.Application) error {
	star := common.GetState(ctx)

	refreshToken := form.Get("refresh_token")
	if refreshToken == "" {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing refresh_token")
	}

	current := star.Database.GetSessionByToken(ctx, refreshToken)
	if current == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired refresh token")
	}

	if current.Status != database.SessionStateActive {
		slog.Warn("refresh token replay detected", "session", current.ID, "user", current.UserID)
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This refresh token has been rotated")
	}

	if current.ExpiresAt.Before(time.Now()) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "Unknown or expired refresh token")
	}

	sessionApp := star.Database.GetApplicationByID(ctx, current.ApplicationID)
	if sessionApp == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The application this refresh token was issued to no longer exists")
	}

	if application != nil && application.ID != sessionApp.ID {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This refresh token was not issued to this client")
	}

	if application == nil && sessionApp.Kind == database.ApplicationKindConfidential {
		return invalidClient(ctx, false, "Confidential clients must authenticate")
	}

	scope := []string(current.Scopes)
	requested := strings.Fields(form.Get("scope"))

	if len(requested) > 0 {
		for _, value := range requested {
			if !slices.Contains(current.Scopes, value) {
				return writeTokenError(ctx, 400, Oauth2ErrorInvalidScope, "The requested scope exceeds the scope originally granted")
			}
		}

		scope = requested
	}

	refreshed, err := session.RefreshSession(ctx, star, current, sessionApp.SessionDuration.Duration())
	if err != nil {
		return err
	}

	var permissions []string
	if sessionApp.ID == star.DefaultApp.ID && len(current.Scopes) == 0 {
		for _, userPermission := range star.Database.GetUserPermissions(ctx, current.UserID) {
			permissions = append(permissions, userPermission.Name)
		}
	}

	accessToken, err := token.Create(ctx, star, refreshed, &token.Options{
		Duration:    sessionApp.TokenDuration.Duration(),
		Scope:       scope,
		Permissions: permissions,
	})

	if err != nil {
		return err
	}

	response := tokenResponse{
		AccessToken:  string(accessToken),
		TokenType:    "Bearer",
		ExpiresIn:    int64(sessionApp.TokenDuration.Duration().Seconds()),
		RefreshToken: refreshed.Token,
		Scope:        strings.Join(scope, " "),
	}

	if slices.Contains(scope, openid.ScopeOpenID) {
		user := star.Database.GetUserByID(ctx, refreshed.UserID)
		if user == nil {
			return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this refresh token was issued for no longer exists")
		}

		claims := openid.UserClaims(ctx, star.Database, user, scope)

		idToken, err := token.CreateIDToken(star, refreshed, sessionApp, claims, nil)
		if err != nil {
			return err
		}

		response.IDToken = string(idToken)
	}

	return writeTokenResponse(ctx, response)
}
