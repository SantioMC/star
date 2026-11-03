package keypair

import (
	"context"
	"slices"
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func Rotate(ctx context.Context, star *state.StarServer) error {
	_, err := rotate(ctx, star, database.KeyAlgorithms, nil)
	return err
}

func RotateExpiring(ctx context.Context, star *state.StarServer) (bool, error) {
	keys := star.SigningKeys.Load()
	if keys == nil {
		return false, nil
	}

	threshold := time.Now().Add(time.Hour)

	expiring := []database.KeyAlgorithm{}
	for _, algorithm := range database.KeyAlgorithms {
		healthy := slices.ContainsFunc(*keys, func(key database.SigningKey) bool {
			return key.Status == database.SigningKeyStatusActive &&
				key.Algorithm == algorithm &&
				(key.ExpiresAt == nil || key.ExpiresAt.After(threshold))
		})

		if !healthy {
			expiring = append(expiring, algorithm)
		}
	}

	if len(expiring) == 0 {
		return false, nil
	}

	return rotate(ctx, star, expiring, &threshold)
}

func rotate(
	ctx context.Context,
	star *state.StarServer,
	algorithms []database.KeyAlgorithm,
	expiringBefore *time.Time,
) (bool, error) {
	longestToken, err := star.Database.GetLongestTokenDuration(ctx)
	if err != nil {
		return false, err
	}

	nextKeys := make([]*database.SigningKey, len(algorithms))
	for idx, algorithm := range algorithms {
		next, err := NewKeypair(algorithm, database.SigningKeyStatusNext, nil)
		if err != nil {
			return false, err
		}

		nextKeys[idx] = next
	}

	now := time.Now()

	return star.Database.RotateSigningKeys(
		ctx,
		nextKeys,
		now.Add(max(longestToken, time.Hour)),
		now.Add(time.Hour*24*7),
		expiringBefore,
	)
}
