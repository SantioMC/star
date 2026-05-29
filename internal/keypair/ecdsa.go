package keypair

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/internal/database"
)

func NewECDSAKeypair(
	status database.SigningKeyStatus,
	duration *time.Duration,
) (*database.SigningKey, error) {
	keypair, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	privateKey, err := MarshalPrivateKey(keypair)
	if err != nil {
		return nil, err
	}

	publicKey, err := MarshalPublicKey(&keypair.PublicKey)
	if err != nil {
		return nil, err
	}

	var expiresAt time.Time
	if duration != nil {
		expiresAt = time.Now().Add(*duration)
	}

	return &database.SigningKey{
		ID:         database.CreateID(database.SigKey),
		Status:     status,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Algorithm:  database.KeyAlgorithm(jose.ES512),
		CreatedAt:  time.Now(),
		ExpiresAt:  &expiresAt,
	}, nil
}
