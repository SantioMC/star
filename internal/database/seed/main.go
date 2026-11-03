package seed

import (
	"context"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/state"
)

func Seed(ctx context.Context, star *state.StarServer) error {
	started := time.Now()

	if err := seedSettings(ctx, star); err != nil {
		return err
	}

	if err := seedBranding(ctx, star); err != nil {
		return err
	}

	if err := seedKeys(ctx, star); err != nil {
		return err
	}

	if err := seedRBAC(ctx, star); err != nil {
		return err
	}

	if err := seedApplication(ctx, star); err != nil {
		return err
	}

	if err := seedUser(ctx, star); err != nil {
		return err
	}

	elapsed := time.Since(started)
	if elapsed > time.Second {
		slog.Warn("seeding took longer than expected, this should likely be looked into", "elapsed", elapsed)
	}

	return nil
}
