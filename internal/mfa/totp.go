package mfa

import (
	"crypto/rand"
	"errors"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tubblelive/star/internal/state"
)

func CreateTotp(star *state.StarServer, userId string) (string, error) {
	value, err := totp.Generate(totp.GenerateOpts{
		SecretSize:  32,
		Issuer:      star.Config.ServerURL,
		AccountName: userId,
		Period:      30,
		Digits:      6,
		Algorithm:   otp.AlgorithmSHA1,
		Rand:        rand.Reader,
	})

	if err != nil {
		return "", err
	}

	return value.String(), nil
}

func VerifyTotp(url string, code string) error {
	key, err := otp.NewKeyFromURL(url)
	if err != nil {
		return err
	}

	if !totp.Validate(code, key.Secret()) {
		return errors.New("invalid totp code provided")
	}

	return nil
}
