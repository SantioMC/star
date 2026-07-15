package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupAuthCodes(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadAuthCodes(ctx)
	if err != nil {
		slog.Error("failed to delete old authorization codes", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted old authorization codes", "amount", deleted)
	}
}
