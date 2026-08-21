package seed

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/state"
)

func seedApplication(ctx context.Context, star *state.StarServer) error {
	defaultStarApplication := &database.Application{
		ID:              database.CreateID(database.ApplicationKey),
		Kind:            database.ApplicationKindPublic,
		Name:            "star",
		Namespace:       openid.StarNamespace,
		ApplicationURL:  star.Config.ServerURL,
		ClientID:        "star",
		ClientSecret:    nil,
		Internal:        true,
		TokenDuration:   models.Seconds(time.Hour),
		SessionDuration: models.Seconds(time.Hour * 24 * 7),
		RedirectURIs:    pq.StringArray{star.Config.ServerURL},
		DevelopmentMode: false,
		CreatedAt:       time.Now(),
	}

	application := star.Database.GetApplicationByName(ctx, "star")
	if application != nil {
		if !application.Internal {
			// This should never happen, this means someone went in the db and changed this, to respect the confgiuration of the organization, we will not overwrite this app anymore - however this is super unsupported and dangerous
			return nil
		}

		defaultStarApplication.ID = application.ID
		defaultStarApplication.CreatedAt = application.CreatedAt

		return star.Database.UpdateApplication(ctx, defaultStarApplication)
	}

	return star.Database.CreateApplication(ctx, defaultStarApplication)
}
