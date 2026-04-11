package security

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func Report(state *state.StarServer, ctx context.Context, alert *database.SecurityAlert) {
	if err := state.Database.CreateSecurityAlert(ctx, alert); err != nil {
		slog.Error("failed to create security alert", "err", err)
	} else {
		slog.Warn("security alert fired", "kind", alert.Kind, "severity", alert.Severity, "message", alert.Message)
	}
}
