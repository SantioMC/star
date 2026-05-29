package cache

import (
	"context"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func StartTicker(state *state.StarServer) {
	ticker := time.NewTicker(time.Minute)

	for {
		<-ticker.C
		UpdateCache(state)
	}
}

func UpdateCache(state *state.StarServer) {
	ctx := context.Background()
	UpdateKeys(ctx, state)
	UpdateSettings(ctx, state)
	UpdateApplication(ctx, state)
}
