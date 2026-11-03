package cache

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func UpdateSettings(ctx context.Context, state *state.StarServer) {
	settings, err := state.Database.GetSettings(ctx)
	if err != nil {
		slog.Error("failed to update settings", "err", err)
		return
	}

	state.Settings.Store(settings)
}
