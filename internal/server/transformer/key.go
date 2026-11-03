package transformer

import (
	"crypto"
	"encoding/base64"
	"time"

	"github.com/tubblelive/star/internal/database"
)

type SigningKey struct {
	ID          string                    `json:"id"`
	Status      database.SigningKeyStatus `json:"status"`
	Algorithm   database.KeyAlgorithm     `json:"algorithm"`
	Fingerprint string                    `json:"fingerprint"`
	CreatedAt   time.Time                 `json:"created_at"`
	ExpiresAt   *time.Time                `json:"expires_at"`
}

func TransformSigningKey(key database.SigningKey) (SigningKey, error) {
	jwk, err := key.GetJWK()
	if err != nil {
		return SigningKey{}, err
	}

	thumbprint, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return SigningKey{}, err
	}

	return SigningKey{
		ID:          key.ID,
		Status:      key.Status,
		Algorithm:   key.Algorithm,
		Fingerprint: base64.RawURLEncoding.EncodeToString(thumbprint),
		CreatedAt:   key.CreatedAt,
		ExpiresAt:   key.ExpiresAt,
	}, nil
}

func TransformSigningKeys(keys []database.SigningKey) ([]SigningKey, error) {
	serialized := []SigningKey{}

	for _, value := range keys {
		key, err := TransformSigningKey(value)
		if err != nil {
			return nil, err
		}

		serialized = append(serialized, key)
	}

	return serialized, nil
}
