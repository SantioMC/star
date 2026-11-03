package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

const PasskeyKey = "pk"

type Passkey struct {
	ID           string                               `json:"id"`
	UserID       string                               `json:"user_id" db:"user_id"`
	Name         string                               `json:"name"`
	CredentialID []byte                               `json:"credential_id" db:"credential_id"`
	Credential   models.JSONData[webauthn.Credential] `json:"credential"`
	LastUsedAt   *time.Time                           `json:"last_used_at" db:"last_used_at"`
	CreatedAt    time.Time                            `json:"created_at" db:"created_at"`
}

func (db *Database) GetUserPasskeys(ctx context.Context, userID string) ([]Passkey, error) {
	return list[Passkey](
		ctx,
		db.DB(),
		`SELECT * FROM passkeys WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
}

func (db *Database) GetUserPasskeyByID(ctx context.Context, userID string, id string) *Passkey {
	passkey, err := sqlx.OneContext[Passkey](
		ctx,
		db.DB(),
		`SELECT * FROM passkeys WHERE id = $1 AND user_id = $2 LIMIT 1`,
		id,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get passkey by id", "err", err)
		return nil
	}

	return &passkey
}

func (db *Database) PasskeyExists(ctx context.Context, credentialID []byte) (bool, error) {
	return sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`SELECT EXISTS (SELECT 1 FROM passkeys WHERE credential_id = $1)`,
		credentialID,
	)
}

func (db *Database) CountUserPasskeys(ctx context.Context, userID string) (int, error) {
	return sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM passkeys WHERE user_id = $1`,
		userID,
	)
}

func (db *Database) CreatePasskey(ctx context.Context, passkey *Passkey) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO passkeys (
			id, user_id, name, credential_id, credential, last_used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		passkey.ID,
		passkey.UserID,
		passkey.Name,
		passkey.CredentialID,
		passkey.Credential,
		passkey.LastUsedAt,
		passkey.CreatedAt,
	)

	return err
}

func (db *Database) RenamePasskey(ctx context.Context, passkey *Passkey) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE passkeys SET name = $2 WHERE id = $1`,
		passkey.ID,
		passkey.Name,
	)

	return err
}

func (db *Database) UpdatePasskeyUsage(ctx context.Context, passkey *Passkey) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE passkeys SET credential = $2, last_used_at = $3 WHERE id = $1`,
		passkey.ID,
		passkey.Credential,
		passkey.LastUsedAt,
	)

	return err
}

func (db *Database) DeletePasskeyByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM passkeys WHERE id = $1`,
		id,
	)

	return err
}
