package session

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
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
	if current.ExpiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	if current.Status != database.SessionStateActive {
		if err := server.Database.DeleteUserSessions(ctx, current.UserID); err != nil {
			slog.Error("failed to invalidate user sessions due to security risk", "user", current.UserID)
		}

		security.Report(server, ctx, &database.SecurityAlert{
			Kind:        database.SecurityAlertASessionReused,
			Severity:    database.SecurityAlertSeverityHigh,
			Message:     "The user's session was rotated multiple times, all user sessions have been automatically revoked",
			UserID:      &current.UserID,
			Application: &current.ApplicationID,
			Data: map[string]any{
				"session":    current.ID,
				"ip_address": ctx.IP(),
				"user_agent": ctx.UserAgent(),
			},
		})

		return nil, ErrSessionRotated
	}

	current.Status = database.SessionStateRotated
	if err := server.Database.UpdateSession(ctx, current); err != nil {
		return nil, err
	}

	session := buildSession(
		ctx,
		server,
		current.UserID,
		current.AccountID,
		current.ApplicationID,
		current.Scopes,
		duration,
		current.AuthenticatedAt,
	)

	if err := server.Database.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}
