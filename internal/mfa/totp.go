package mfa

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

func CreateTotp(star *state.StarServer, user *database.User) (string, string, error) {
	value, err := totp.Generate(totp.GenerateOpts{
		SecretSize:  32,
		Issuer:      "star",
		AccountName: user.Name,
		Period:      30,
		Digits:      6,
		Algorithm:   otp.AlgorithmSHA1,
		Rand:        rand.Reader,
	})

	if err != nil {
		return "", "", err
	}

	return value.Secret(), value.String(), nil
}

var ErrInvalidTotp = errors.New("invalid totp code provided")

func VerifyTotp(ctx context.Context, star *state.StarServer, record *database.Totp, code string) error {
	now := time.Now()

	for _, skew := range []time.Duration{0, -30 * time.Second, 30 * time.Second} {
		at := now.Add(skew)
		expected, err := totp.GenerateCodeCustom(record.Secret, at, totp.ValidateOpts{
			Period:    30,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})

		if err != nil {
			return err
		}

		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) != 1 {
			continue
		}

		used, err := star.Database.UseTotpStep(ctx, record.ID, at.Unix()/30)
		if err != nil {
			return err
		} else if !used {
			return ErrInvalidTotp
		}

		return nil
	}

	return ErrInvalidTotp
}
