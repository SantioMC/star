package totp

import (
	"crypto/rand"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func Create(userId string) (string, error) {
	value, err := totp.Generate(totp.GenerateOpts{
		SecretSize:  32,
		Issuer:      "star",
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
