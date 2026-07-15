package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupSessions(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadSessions(ctx)
	if err != nil {
		slog.Error("failed to delete old sessions", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted old sessions", "amount", deleted)
	}
}
