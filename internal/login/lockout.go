package login

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/security"
	"github.com/tubblelive/star/internal/state"
)

var ErrLoginLocked = fiber.NewError(429, "Too many failed attempts, please try again later")

func ReserveAttempt(ctx fiber.Ctx, star *state.StarServer, userID string) (*database.LoginAttempts, error) {
	policy := star.Settings.Load().GetLockoutPolicy()
	attempt, err := star.Database.ReserveLoginAttempt(ctx, userID, policy)

	if errors.Is(err, database.ErrLoginLocked) {
		return nil, ErrLoginLocked
	}

	return attempt, err
}

func ReportFailure(ctx fiber.Ctx, star *state.StarServer, attempt *database.LoginAttempts, kind database.SecurityAlertKind) {
	if attempt.LockedUntil == nil {
		return
	}

	audit.Log(
		ctx,
		audit.LoginLocked,
		audit.ByID(attempt.UserID),
		audit.With("attempts", attempt.Attempts),
		audit.With("locked_until", attempt.LockedUntil),
		audit.With("reason", string(kind)),
	)

	security.Report(star, ctx, &database.SecurityAlert{
		Kind:     kind,
		Severity: database.SecurityAlertSeverityHigh,
		Message:  "The user failed to sign in repeatedly, further attempts are blocked until the lockout expires",
		UserID:   &attempt.UserID,
		Data: map[string]any{
			"attempts":     attempt.Attempts,
			"locked_until": attempt.LockedUntil,
			"ip_address":   star.ClientIP(ctx),
			"user_agent":   ctx.UserAgent(),
		},
	})
}
