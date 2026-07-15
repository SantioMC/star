package mfa

import (
	"context"

	"github.com/tubblelive/star/internal/state"
)

type MFAMethod string

const (
	MFATotp MFAMethod = "totp"
)

func GetMFA(ctx context.Context, star *state.StarServer, userId string) ([]MFAMethod, error) {
	methods := []MFAMethod{}

	totp, err := star.Database.IsTotpEnabled(ctx, userId)
	if err != nil {
		return nil, err
	} else if totp {
		methods = append(methods, MFATotp)
	}

	return methods, nil
}
