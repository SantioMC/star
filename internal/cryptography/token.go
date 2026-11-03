package cryptography

import (
	"crypto/rand"
	"math/big"
)

func CreateSecureString(length int) string {
	return CreateSecureStringFrom("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", length)
}

func CreateSecureStringFrom(alphabet string, length int) string {
	ret := make([]byte, length)
	limit := big.NewInt(int64(len(alphabet)))

	for i := range length {
		num, err := rand.Int(rand.Reader, limit)
		if err != nil {
			panic(err)
		}

		ret[i] = alphabet[num.Int64()]
	}

	return string(ret)
}
