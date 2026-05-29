package cache

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/state"
)

func UpdateApplication(ctx context.Context, state *state.StarServer) {
	app := state.Database.GetApplicationByName(ctx, "star")
	if app == nil {
		slog.Error("failed to get built-in application, requests will fail")
		return
	}

	state.DefaultApp = app
}
