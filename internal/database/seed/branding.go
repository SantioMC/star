package seed

import (
	"context"

	"github.com/tubblelive/star/internal/state"
)

func seedBranding(ctx context.Context, star *state.StarServer) error {
	if star.Database.BrandingExists(ctx) {
		return nil
	}

	return star.Database.SaveDefaultBranding(ctx)
}
