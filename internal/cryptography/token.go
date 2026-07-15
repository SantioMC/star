package cryptography

import (
	"crypto/rand"
	"math/big"
)

var alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func CreateSecureString(length int) string {
	ret := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}

		ret[i] = alphabet[num.Int64()]
	}

	return string(ret)
}
