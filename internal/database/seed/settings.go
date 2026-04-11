package seed

import (
	"context"

	"github.com/tubblelive/star/internal/state"
)

func seedSettings(ctx context.Context, star *state.StarServer) error {
	exists := star.Database.SettingsExists(ctx)
	if exists {
		return nil
	}

	return star.Database.SaveDefaultSettings(ctx)
}
