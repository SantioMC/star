package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupDeviceAuthorizations(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadDeviceAuthorizations(ctx)
	if err != nil {
		slog.Error("failed to delete old device authorizations", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted old device authorizations", "amount", deleted)
	}
}
