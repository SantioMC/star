package cryptography

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

func HMAC(value string, key [32]byte) string {
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte(value))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func Sign(purpose string, payload []byte, key [32]byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + HMAC(purpose+":"+encoded, key)
}

func Verify(purpose string, signed string, key [32]byte) ([]byte, bool) {
	encoded, signature, found := strings.Cut(signed, ".")
	if !found {
		return nil, false
	}

	expected := HMAC(purpose+":"+encoded, key)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) != 1 {
		return nil, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	return payload, err == nil
}
