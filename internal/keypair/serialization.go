package keypair

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
)

func MarshalPrivateKey(key crypto.Signer) (string, error) {
	block, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	data := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: block})
	return string(data), nil
}

func MarshalPublicKey(key crypto.PublicKey) (string, error) {
	block, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	data := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: block})
	return string(data), nil
}
