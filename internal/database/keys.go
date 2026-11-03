package database

import (
	"context"
	"crypto"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const SigKey = "sig"

type SigningKeyStatus string
type KeyAlgorithm jose.SignatureAlgorithm

const (
	SigningKeyStatusActive   SigningKeyStatus = "active"
	SigningKeyStatusNext     SigningKeyStatus = "next"
	SigningKeyStatusRetiring SigningKeyStatus = "retiring"
)

const (
	KeyAlgorithmES512 = KeyAlgorithm(jose.ES512)
	KeyAlgorithmRS256 = KeyAlgorithm(jose.RS256)
)

var KeyAlgorithms = []KeyAlgorithm{KeyAlgorithmES512, KeyAlgorithmRS256}

func SignatureAlgorithms() []jose.SignatureAlgorithm {
	algorithms := make([]jose.SignatureAlgorithm, len(KeyAlgorithms))
	for idx, algorithm := range KeyAlgorithms {
		algorithms[idx] = jose.SignatureAlgorithm(algorithm)
	}

	return algorithms
}

type SigningKey struct {
	ID         string           `json:"id"`
	Status     SigningKeyStatus `json:"status"`
	PublicKey  string           `json:"public_key" db:"public_key"`
	PrivateKey string           `json:"-" db:"private_key"`
	Algorithm  KeyAlgorithm     `json:"algorithm" db:"algorithm"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time       `json:"expires_at" db:"expires_at"`
}

func (k *SigningKey) IsUsable() bool {
	return k.Status == SigningKeyStatusActive && (k.ExpiresAt == nil || k.ExpiresAt.After(time.Now()))
}

func (k *SigningKey) GetPrivateKey() (crypto.Signer, error) {
	block, _ := pem.Decode([]byte(k.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key of type %T can not sign", key)
	}

	return signer, nil
}

func (k *SigningKey) GetPublicKey() (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(k.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	return x509.ParsePKIXPublicKey(block.Bytes)
}

func (k *SigningKey) GetJWK() (jose.JSONWebKey, error) {
	publicKey, err := k.GetPublicKey()
	if err != nil {
		return jose.JSONWebKey{}, err
	}

	jwk := jose.JSONWebKey{
		Key:       publicKey,
		Use:       "sig",
		Algorithm: string(k.Algorithm),
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

	if signingKey.PublicKey, err = cryptography.DecryptAES(signingKey.PublicKey, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt public key", "err", err)
		return nil
	}

	if signingKey.PrivateKey, err = cryptography.DecryptAES(signingKey.PrivateKey, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt private key", "err", err)
		return nil
	}

	return &signingKey
}

func (db *Database) GetAllSigningKeys(ctx context.Context) []SigningKey {
	signingKeys, err := list[SigningKey](
		ctx,
		db.DB(),
		`SELECT * FROM signing_keys ORDER BY created_at ASC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []SigningKey{}
	} else if err != nil {
		slog.Error("failed to get signing keys", "err", err)
		return nil
	}

	for idx, signingKey := range signingKeys {
		if signingKeys[idx].PublicKey, err = cryptography.DecryptAES(signingKey.PublicKey, db.config.AppToken); err != nil {
			slog.Error("failed to decrypt public key", "err", err)
			return nil
		}

		if signingKeys[idx].PrivateKey, err = cryptography.DecryptAES(signingKey.PrivateKey, db.config.AppToken); err != nil {
			slog.Error("failed to decrypt private key", "err", err)
			return nil
		}
	}

	return signingKeys
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
	return db.insertSigningKey(ctx, db.DB(), signingKey)
}

func (db *Database) insertSigningKey(ctx context.Context, executor sqlx.ExecerContext, signingKey *SigningKey) error {
	encryptedPublicKey, err := cryptography.EncryptAES(signingKey.PublicKey, db.config.AppToken)
	if err != nil {
		return err
	}

	encryptedPrivateKey, err := cryptography.EncryptAES(signingKey.PrivateKey, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = executor.ExecContext(
		ctx,
		`
		INSERT INTO signing_keys (id, status, public_key, private_key, algorithm, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		signingKey.ID,
		signingKey.Status,
		encryptedPublicKey,
		encryptedPrivateKey,
		signingKey.Algorithm,
		signingKey.CreatedAt,
		signingKey.ExpiresAt,
	)

	return err
}

func (db *Database) RotateSigningKeys(
	ctx context.Context,
	nextKeys []*SigningKey,
	retireAt time.Time,
	activeUntil time.Time,
	expiringBefore *time.Time,
) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	sorted := slices.SortedFunc(slices.Values(nextKeys), func(a, b *SigningKey) int {
		return strings.Compare(string(a.Algorithm), string(b.Algorithm))
	})

	rotated := false
	for _, next := range sorted {
		changed, err := db.rotateSigningKey(ctx, tx, next, retireAt, activeUntil, expiringBefore)
		if err != nil {
			return false, err
		}

		rotated = rotated || changed
	}

	return rotated, tx.Commit()
}

func (db *Database) rotateSigningKey(
	ctx context.Context,
	tx *sqlx.Tx,
	next *SigningKey,
	retireAt time.Time,
	activeUntil time.Time,
	expiringBefore *time.Time,
) (bool, error) {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, next.Algorithm); err != nil {
		return false, err
	}

	retired, err := tx.ExecContext(
		ctx,
		`
		UPDATE signing_keys SET status = $1, expires_at = $2
		WHERE status = $3 AND algorithm = $4 AND ($5::TIMESTAMPTZ IS NULL OR expires_at < $5)
		`,
		SigningKeyStatusRetiring,
		retireAt,
		SigningKeyStatusActive,
		next.Algorithm,
		expiringBefore,
	)

	if err != nil {
		return false, err
	}

	if count, err := retired.RowsAffected(); err != nil {
		return false, err
	} else if count == 0 && expiringBefore != nil {
		return false, nil
	}

	promoted, err := tx.ExecContext(
		ctx,
		`UPDATE signing_keys SET status = $1, expires_at = $2 WHERE status = $3 AND algorithm = $4`,
		SigningKeyStatusActive,
		activeUntil,
		SigningKeyStatusNext,
		next.Algorithm,
	)

	if err != nil {
		return false, err
	}

	if count, err := promoted.RowsAffected(); err != nil {
		return false, err
	} else if count != 1 {
		return false, fmt.Errorf("expected exactly one next %s signing key to promote, found %d", next.Algorithm, count)
	}

	if err := db.insertSigningKey(ctx, tx, next); err != nil {
		return false, err
	}

	return true, nil
}

func (db *Database) DeleteRetiringSigningKey(ctx context.Context, id string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM signing_keys WHERE id = $1 AND status = $2`,
		id,
		SigningKeyStatusRetiring,
	)

	if err != nil {
		return false, err
	}

	deleted, err := result.RowsAffected()
	return deleted > 0, err
}

func (db *Database) GetSigningKeysCount(ctx context.Context, algorithm KeyAlgorithm) (int, error) {
	count, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM signing_keys WHERE algorithm = $1`,
		algorithm,
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
