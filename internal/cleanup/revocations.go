package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupTokenRevocations(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteExpiredTokenRevocations(ctx)
	if err != nil {
		slog.Error("failed to delete expired token revocations", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted expired token revocations", "amount", deleted)
	}
}
