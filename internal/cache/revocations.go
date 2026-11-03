package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func UpdateRevocations(ctx context.Context, state *state.StarServer) {
	since := state.Revocations.Watermark().Add(-time.Minute)

	revocations, err := state.Database.GetTokenRevocationsSince(ctx, since)
	if err != nil {
		slog.Error("failed to update token revocations", "err", err)
		return
	}

	state.Revocations.Apply(revocations)
}
