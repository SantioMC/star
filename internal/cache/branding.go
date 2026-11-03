package cache

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func UpdateBranding(ctx context.Context, state *state.StarServer) {
	branding, err := state.Database.GetBranding(ctx)
	if err != nil {
		slog.Error("failed to update branding", "err", err)
		return
	}

	state.Branding.Store(branding)
}
