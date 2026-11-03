package session

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/security"
	"github.com/tubblelive/star/internal/state"
)

var ErrSessionExpired = errors.New("this session has expired")
var ErrSessionRotated = errors.New("this session has already been rotated")

func RefreshSession(
	ctx fiber.Ctx,
	server *state.StarServer,
	current *database.Session,
	duration time.Duration,
) (*database.Session, error) {
	if current.Status != database.SessionStateActive {
		RevokeReplayedSession(ctx, server, current)
		return nil, ErrSessionRotated
	}

	if current.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	session := buildSession(
		ctx,
		server,
		current.FamilyID,
		current.UserID,
		current.AccountID,
		current.ApplicationID,
		current.Scopes,
		duration,
		current.AuthenticatedAt,
	)

	rotated, err := server.Database.RotateSession(ctx, current, session)
	if err != nil {
		return nil, err
	} else if !rotated {
		RevokeReplayedSession(ctx, server, current)
		return nil, ErrSessionRotated
	}

	return session, nil
}

func RevokeReplayedSession(ctx fiber.Ctx, server *state.StarServer, current *database.Session) {
	err := server.Database.RevokeSessionFamily(ctx, current.UserID, current.FamilyID, database.RevocationReasonTokenReplayed)
	if err != nil {
		slog.Error("failed to revoke replayed session family", "err", err, "user", current.UserID, "family", current.FamilyID)
	}

	cache.UpdateRevocations(ctx, server)

	security.Report(server, ctx, &database.SecurityAlert{
		Kind:        database.SecurityAlertASessionReused,
		Severity:    database.SecurityAlertSeverityHigh,
		Message:     "A rotated refresh token was used again, the session it belongs to has been revoked",
		UserID:      &current.UserID,
		Application: &current.ApplicationID,
		Data: map[string]any{
			"session":    current.FamilyID,
			"ip_address": server.ClientIP(ctx),
			"user_agent": ctx.UserAgent(),
		},
	})
}
