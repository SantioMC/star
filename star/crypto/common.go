package crypto

import (
	"crypto/rand"
	"math/big"
)

var alphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@$^~")

func CreateRand(length int) []byte {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		panic("rand.Read() returns an error which is said to not be possible")
	}

	return salt
}

func CreateRandString(length int) string {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))-1))
		if err != nil {
			panic("rand.Int() returns an error which should've been impossible")
		}

		ret[i] = alphabet[num.Int64()]
	}

	return string(ret)
}
