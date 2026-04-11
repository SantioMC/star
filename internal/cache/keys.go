package cache

import (
	"context"

	"github.com/tubblelive/star/internal/state"
)

func UpdateKeys(ctx context.Context, state *state.StarServer) {
	keys := state.Database.GetAllSigningKeys(ctx)
	if keys == nil {
		return
	}

	state.SigningKeys = &keys
}
