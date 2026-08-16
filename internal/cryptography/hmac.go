package cryptography

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func HMAC(value string, key [32]byte) string {
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte(value))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
