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

	auditTicker := time.NewTicker(time.Hour)
	PrepareAuditLogs(context.Background(), state)

	go func() {
		CleanupAuditLogs(context.Background(), state)

		for {
			<-auditTicker.C
			PrepareAuditLogs(context.Background(), state)
			CleanupAuditLogs(context.Background(), state)
		}
	}()
}

func Cleanup(state *state.StarServer) {
	ctx := context.Background()
	CleanupSessions(ctx, state)
	CleanupAuthCodes(ctx, state)
	CleanupImpersonations(ctx, state)
	CleanupLoginFlows(ctx, state)
	CleanupPasskeyChallenges(ctx, state)
	CleanupPasswordResets(ctx, state)
	CleanupExpiredInvites(ctx, state)
	CleanupSigningKeys(ctx, state)
	CleanupTokenRevocations(ctx, state)
}
