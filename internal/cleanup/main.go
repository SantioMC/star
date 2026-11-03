package cleanup

import (
	"context"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func StartTicker(state *state.StarServer) {
	ticker := time.NewTicker(time.Minute)
	Cleanup(state)

	go func() {
		for {
			<-ticker.C
			Cleanup(state)
		}
	}()
}

func Cleanup(state *state.StarServer) {
	ctx := context.Background()
	CleanupSessions(ctx, state)
	CleanupAuthCodes(ctx, state)
	CleanupLoginFlows(ctx, state)
	CleanupPasskeyChallenges(ctx, state)
	CleanupSigningKeys(ctx, state)
	CleanupTokenRevocations(ctx, state)
}
