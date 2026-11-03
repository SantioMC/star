package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func CleanupPasskeyChallenges(ctx context.Context, star *state.StarServer) {
	deleted, err := star.Database.DeleteDeadPasskeyChallenges(ctx)
	if err != nil {
		slog.Error("failed to delete old passkey challenges", "err", err)
		return
	}

	if deleted != 0 {
		slog.Debug("deleted old passkey challenges", "amount", deleted)
	}
}
