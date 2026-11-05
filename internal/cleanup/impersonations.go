package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupImpersonations(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadImpersonations(ctx)
	if err != nil {
		slog.Error("failed to delete expired impersonations", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted expired impersonations", "amount", deleted)
	}
}
