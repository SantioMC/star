package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/keypair"
	"github.com/tubblelive/star/internal/state"
)

func CleanupSigningKeys(ctx context.Context, star *state.StarServer) {
	if err := star.Database.DeletedRetiredSigningKeys(ctx); err != nil {
		slog.Error("failed to delete retired signing keys", "err", err)
	}

	rotated, err := keypair.RotateExpiring(ctx, star)
	if err != nil {
		slog.Error("failed to rotate expiring signing keys", "err", err)
	}

	if rotated {
		slog.Info("rotated expiring signing keys")
		cache.UpdateKeys(ctx, star)
	}
}
