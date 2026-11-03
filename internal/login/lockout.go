package login

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/security"
	"github.com/tubblelive/star/internal/state"
)

var ErrLoginLocked = fiber.NewError(429, "Too many failed attempts, please try again later")

func ReserveAttempt(ctx fiber.Ctx, star *state.StarServer, userID string) (int, error) {
	attempt, err := star.Database.ReserveLoginAttempt(ctx, userID)
	if errors.Is(err, database.ErrLoginLocked) {
		return 0, ErrLoginLocked
	}

	return attempt, err
}

func ReportFailure(ctx fiber.Ctx, star *state.StarServer, userID string, attempt int, kind database.SecurityAlertKind) {
	if attempt != 5 {
		return
	}

	security.Report(star, ctx, &database.SecurityAlert{
		Kind:     kind,
		Severity: database.SecurityAlertSeverityHigh,
		Message:  "The user failed to sign in repeatedly, further attempts are being throttled",
		UserID:   &userID,
		Data: map[string]any{
			"attempts":   attempt,
			"ip_address": star.ClientIP(ctx),
			"user_agent": ctx.UserAgent(),
		},
	})
}
