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
	keys, err := star.Database.GetSigningKeysCount(ctx)
	if err != nil {
		return err
	}

	if keys > 0 {
		return nil
	}

	defaultKeyDuration := time.Hour * 24 * 7

	current, err := keypair.NewECDSAKeypair(
		database.SigningKeyStatusActive,
		&defaultKeyDuration,
	)

	if err != nil {
		return err
	}

	if err := star.Database.CreateSigningKey(ctx, current); err != nil {
		return err
	}

	next, err := keypair.NewECDSAKeypair(
		database.SigningKeyStatusNext,
		nil,
	)

	if err != nil {
		return err
	}

	if err := star.Database.CreateSigningKey(ctx, next); err != nil {
		return err
	}

	return nil
}
