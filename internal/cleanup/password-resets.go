package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupPasswordResets(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteExpiredPasswordResets(ctx)
	if err != nil {
		slog.Error("failed to delete expired password resets", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted expired password resets", "amount", deleted)
	}
}
