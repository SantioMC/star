package keypair

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/tubblelive/star/internal/database"
)

func generateKey(algorithm database.KeyAlgorithm) (crypto.Signer, error) {
	switch algorithm {
	case database.KeyAlgorithmES512:
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case database.KeyAlgorithmRS256:
		return rsa.GenerateKey(rand.Reader, 2048)
	default:
		return nil, fmt.Errorf("unsupported signing key algorithm: %s", algorithm)
	}
}

func NewKeypair(
	algorithm database.KeyAlgorithm,
	status database.SigningKeyStatus,
	duration *time.Duration,
) (*database.SigningKey, error) {
	keypair, err := generateKey(algorithm)
	if err != nil {
		return nil, err
	}

	privateKey, err := MarshalPrivateKey(keypair)
	if err != nil {
		return nil, err
	}

	publicKey, err := MarshalPublicKey(keypair.Public())
	if err != nil {
		return nil, err
	}

	var expiresAt *time.Time
	if duration != nil {
		expiresAt = new(time.Now().Add(*duration))
	}

	return &database.SigningKey{
		ID:         database.CreateID(database.SigKey),
		Status:     status,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Algorithm:  algorithm,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
	}, nil
}
