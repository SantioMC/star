package oauth

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/birthday"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/security"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

type issuedGrant struct {
	label           string
	userID          string
	impersonatorID  *string
	scopes          []string
	nonce           *string
	authenticatedAt time.Time
	redeem          func(familyID string) (bool, error)
	replayed        func() error
}

type grantReplay struct {
	kind          database.SecurityAlertKind
	label         string
	message       string
	id            string
	userID        string
	applicationID string
	familyID      *string
}

func redeemGrant(ctx fiber.Ctx, application *database.Application, grant *issuedGrant) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, grant.userID)
	if user == nil {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this "+grant.label+" was issued for no longer exists")
	}

	if !star.Database.UserCanAccessApplication(ctx, user.ID, application) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this "+grant.label+" was issued for can not access this application")
	}

	resolution := openid.ResolveScopes(ctx, star.Database, application, grant.scopes)
	if len(resolution.Granted) == 0 {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidScope, "None of the scopes granted by this "+grant.label+" are available anymore")
	}

	granted := resolution.Names()

	newSession, err := session.CreateSession(ctx, star, &session.Options{
		User:            user,
		Application:     application,
		ImpersonatorID:  grant.impersonatorID,
		Scopes:          granted,
		Duration:        application.SessionDuration.Duration(),
		AuthenticatedAt: grant.authenticatedAt,
	})

	if errors.Is(err, session.ErrUserSuspended) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this "+grant.label+" was issued for is suspended")
	} else if errors.Is(err, birthday.ErrAgeRestricted) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The user this "+grant.label+" was issued for does not meet the age requirement")
	} else if errors.Is(err, session.ErrImpersonationDenied) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "The impersonator of this "+grant.label+" is no longer permitted to impersonate this user")
	} else if err != nil {
		return err
	}

	redeemed, err := grant.redeem(newSession.FamilyID)
	if err != nil {
		return err
	}

	if !redeemed {
		if err := star.Database.DeleteSessionByID(ctx, newSession.ID); err != nil {
			slog.Error("failed to delete session after losing "+grant.label+" redemption race", "err", err, "session", newSession.ID)
		}

		return grant.replayed()
	}

	claims := openid.UserClaims(ctx, star.Database, user, granted, star.Settings.Load().TrackBirthdays)
	accessToken, err := token.Create(ctx, star, newSession, &token.Options{
		Application: application,
		Duration:    application.TokenDuration.Duration(),
		Scope:       granted,
		Audience:    resolution.Audiences(application.ID, star.DefaultApp.Load().ID),
		Claims:      claims,
	})

	if errors.Is(err, token.ErrSessionRevoked) {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This "+grant.label+" has been revoked")
	} else if err != nil {
		return err
	}

	response := tokenResponse{
		AccessToken: string(accessToken),
		TokenType:   "Bearer",
		ExpiresIn:   int64(application.TokenDuration.Duration().Seconds()),
		Scope:       strings.Join(granted, " "),
	}

	if openid.HasScope(granted, openid.ScopeOfflineAccess) {
		response.RefreshToken = newSession.Token
	}

	if openid.HasScope(granted, openid.ScopeOpenID) {
		idToken, err := token.CreateIDToken(star, newSession, application, claims, grant.nonce)
		if err != nil {
			return err
		}

		response.IDToken = string(idToken)
	}

	return writeTokenResponse(ctx, response)
}

func revokeReplayedGrant(ctx fiber.Ctx, star *state.StarServer, replay *grantReplay) error {
	slog.Warn(replay.label+" replay detected, revoking the previously issued session", "code", replay.id, "application", replay.applicationID)

	if replay.familyID != nil {
		err := star.Database.RevokeSessionFamily(ctx, replay.userID, *replay.familyID, database.RevocationReasonCodeReplayed)
		if err != nil {
			slog.Error("failed to revoke session for replayed "+replay.label, "err", err, "session", *replay.familyID)
		}

		cache.UpdateRevocations(ctx, star)
	}

	security.Report(star, ctx, &database.SecurityAlert{
		Kind:        replay.kind,
		Severity:    database.SecurityAlertSeverityHigh,
		Message:     replay.message,
		UserID:      &replay.userID,
		Application: &replay.applicationID,
		Data: map[string]any{
			"code":       replay.id,
			"session":    replay.familyID,
			"ip_address": star.ClientIP(ctx),
			"user_agent": ctx.UserAgent(),
		},
	})

	return writeTokenError(ctx, 400, Oauth2ErrorInvalidGrant, "This "+replay.label+" has already been redeemed")
}
