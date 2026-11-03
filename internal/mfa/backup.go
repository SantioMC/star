package mfa

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/tubblelive/star/internal/state"
)

func createBackupCode() string {
	alphabet := "abcdefghjkmnpqrstuvwxyz23456789"
	code := make([]byte, 10)

	for idx := range code {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}

		code[idx] = alphabet[index.Int64()]
	}

	return string(code[:5]) + "-" + string(code[5:])
}

func normalizeBackupCode(code string) string {
	stripped := strings.NewReplacer("-", "", " ", "").Replace(code)
	return strings.ToLower(stripped)
}

func RegenerateBackupCodes(ctx context.Context, star *state.StarServer, userID string) ([]string, error) {
	codes := make([]string, 0, 10)
	normalized := make([]string, 0, 10)

	for range 10 {
		code := createBackupCode()

		codes = append(codes, code)
		normalized = append(normalized, normalizeBackupCode(code))
	}

	if err := star.Database.ReplaceBackupCodes(ctx, userID, normalized); err != nil {
		return nil, err
	}

	return codes, nil
}

func UseBackupCode(ctx context.Context, star *state.StarServer, userID string, code string) (bool, error) {
	return star.Database.ConsumeBackupCode(ctx, userID, normalizeBackupCode(code))
}

func IssueInitialBackupCodes(ctx context.Context, star *state.StarServer, userID string) ([]string, error) {
	remaining, err := star.Database.CountBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	} else if remaining > 0 {
		return nil, nil
	}

	return RegenerateBackupCodes(ctx, star, userID)
}

func RemoveOrphanedBackupCodes(ctx context.Context, star *state.StarServer, userID string) error {
	enabled, err := HasSecondFactor(ctx, star, userID)
	if err != nil || enabled {
		return err
	}

	return star.Database.DeleteBackupCodes(ctx, userID)
}
