package seed

import (
	"context"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/keypair"
	"github.com/tubblelive/star/internal/state"
)

// Create an active signing key and prepare the next one for rotation
func seedKeys(ctx context.Context, star *state.StarServer) error {
	for _, algorithm := range database.KeyAlgorithms {
		if err := seedAlgorithmKeys(ctx, star, algorithm); err != nil {
			return err
		}
	}

	return nil
}

func seedAlgorithmKeys(ctx context.Context, star *state.StarServer, algorithm database.KeyAlgorithm) error {
	keys, err := star.Database.GetSigningKeysCount(ctx, algorithm)
	if err != nil {
		return err
	}

	if keys > 0 {
		return nil
	}

	defaultKeyDuration := time.Hour * 24 * 7

	current, err := keypair.NewKeypair(
		algorithm,
		database.SigningKeyStatusActive,
		&defaultKeyDuration,
	)

	if err != nil {
		return err
	}

	if err := star.Database.CreateSigningKey(ctx, current); err != nil {
		return err
	}

	next, err := keypair.NewKeypair(
		algorithm,
		database.SigningKeyStatusNext,
		nil,
	)

	if err != nil {
		return err
	}

	return star.Database.CreateSigningKey(ctx, next)
}
