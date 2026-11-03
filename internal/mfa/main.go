package mfa

import (
	"context"

	"github.com/tubblelive/star/internal/state"
)

type MFAMethod string

const (
	MFATotp       MFAMethod = "totp"
	MFAPasskey    MFAMethod = "passkey"
	MFABackupCode MFAMethod = "backup_code"
)

func GetMFA(ctx context.Context, star *state.StarServer, userId string) ([]MFAMethod, error) {
	methods := []MFAMethod{}

	totp, err := star.Database.IsTotpEnabled(ctx, userId)
	if err != nil {
		return nil, err
	} else if totp {
		methods = append(methods, MFATotp)
	}

	passkeys, err := star.Database.CountUserPasskeys(ctx, userId)
	if err != nil {
		return nil, err
	} else if passkeys > 0 {
		methods = append(methods, MFAPasskey)
	}

	if len(methods) == 0 {
		return methods, nil
	}

	backupCodes, err := star.Database.CountBackupCodes(ctx, userId)
	if err != nil {
		return nil, err
	} else if backupCodes > 0 {
		methods = append(methods, MFABackupCode)
	}

	return methods, nil
}

func HasSecondFactor(ctx context.Context, star *state.StarServer, userId string) (bool, error) {
	methods, err := GetMFA(ctx, star, userId)
	return len(methods) > 0, err
}
