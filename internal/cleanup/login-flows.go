package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupLoginFlows(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadLoginFlows(ctx)
	if err != nil {
		slog.Error("failed to delete old login flows", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted old login flows", "amount", deleted)
	}
}
