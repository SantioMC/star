package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

func MarshalPrivateKey(key *ecdsa.PrivateKey) (string, error) {
	block, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	data := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: block})
	return string(data), nil
}

func MarshalPublicKey(key *ecdsa.PublicKey) (string, error) {
	block, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	data := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: block})
	return string(data), nil
}