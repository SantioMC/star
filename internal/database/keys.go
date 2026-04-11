package database

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/vinovest/sqlx"
)

type SigningKeyStatus string
type KeyAlgorithm jose.SignatureAlgorithm

const (
	SigningKeyStatusActive   SigningKeyStatus = "active"
	SigningKeyStatusNext     SigningKeyStatus = "next"
	SigningKeyStatusRetiring SigningKeyStatus = "retiring"
)

type SigningKey struct {
	ID         string           `json:"id"`
	Status     SigningKeyStatus `json:"status"`
	PublicKey  string           `json:"public_key" db:"public_key"`
	PrivateKey string           `json:"private_key" db:"private_key"`
	Algorithm  KeyAlgorithm     `json:"algorithm" db:"algorithm"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time       `json:"expires_at" db:"expires_at"`
}

func (k *SigningKey) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(k.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*ecdsa.PrivateKey), nil
}

func (k *SigningKey) GetPublicKey() (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(k.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*ecdsa.PublicKey), nil
}

func (k *SigningKey) GetJWK() (jose.JSONWebKey, error) {
	publicKey, err := k.GetPublicKey()
	if err != nil {
		return jose.JSONWebKey{}, err
	}

	jwk := jose.JSONWebKey{
		Key:       publicKey,
		Use:       "sig",
		Algorithm: string(jose.ES512),
		KeyID:     k.ID,
	}

	return jwk, nil
}

func (db *Database) GetSigningKeyByID(ctx context.Context, id string) *SigningKey {
	signingKey, err := sqlx.OneContext[SigningKey](
		ctx,
		db.DB(),
		`SELECT * FROM signing_keys WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get signing key by id", "err", err)
		return nil
	}

	return &signingKey
}

func (db *Database) GetAllSigningKeys(ctx context.Context) []SigningKey {
	signingKeys, err := sqlx.ListContext[SigningKey](
		ctx,
		db.DB(),
		`SELECT * FROM signing_keys ORDER BY created_at ASC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get signing keys", "err", err)
		return nil
	}

	return signingKeys
}

func (db *Database) GetActiveSigningKey(ctx context.Context) *SigningKey {
	signingKey, err := sqlx.OneContext[SigningKey](
		ctx,
		db.DB(),
		`SELECT * FROM signing_keys WHERE status = $1 AND expires_at > $2 LIMIT 1`,
		SigningKeyStatusActive,
		time.Now(),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get active signing key", "err", err)
		return nil
	}

	return &signingKey
}

func (db *Database) DeletedRetiredSigningKeys(ctx context.Context) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM signing_keys WHERE status = $1 AND expires_at < $2`,
		SigningKeyStatusRetiring,
		time.Now().Add(-time.Hour*24*7),
	)

	return err
}

func (db *Database) CreateSigningKey(ctx context.Context, signingKey *SigningKey) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO signing_keys (id, status, public_key, private_key, algorithm, created_at, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		signingKey.ID,
		signingKey.Status,
		signingKey.PublicKey,
		signingKey.PrivateKey,
		signingKey.Algorithm,
		signingKey.CreatedAt,
		signingKey.ExpiresAt,
	)

	return err
}

func (db *Database) GetSigningKeysCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM signing_keys`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
