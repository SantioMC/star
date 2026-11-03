package cache

import (
	"context"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func StartTicker(state *state.StarServer) {
	ticker := time.NewTicker(time.Second * 15)
	UpdateCache(state)

	go func() {
		for {
			<-ticker.C
			UpdateCache(state)
		}
	}()
}

func UpdateCache(state *state.StarServer) {
	ctx := context.Background()
	UpdateKeys(ctx, state)
	UpdateSettings(ctx, state)
	UpdateApplication(ctx, state)
	UpdateRevocations(ctx, state)
}
