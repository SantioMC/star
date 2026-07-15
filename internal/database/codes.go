package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const AuthCodeKey = "code"

type AuthorizationCode struct {
	ID string `json:"id"`

	UserID      string `json:"user_id"`
	Application string `json:"application"`

	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`

	Scopes []string `json:"scopes"`
	State  *string  `json:"state"`
	Nonce  *string  `json:"nonce"`

	CodeChallenge       *string `json:"code_challenge"`
	CodeChallengeMethod *string `json:"code_challenge_method"`

	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RedeemedAt *time.Time `json:"redeemed_at"`
}

func (db *Database) GetAuthorizationCodeByID(ctx context.Context, id string) *AuthorizationCode {
	authCode, err := sqlx.OneContext[AuthorizationCode](
		ctx,
		db.DB(),
		`SELECT * FROM authorization_codes WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get authorization code by id", "err", err)
		return nil
	}

	if authCode.Code, err = cryptography.DecryptAES(authCode.Code, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt auth code", "err", err)
		return nil
	}

	return &authCode
}

func (db *Database) GetAuthorizationCodeByCode(ctx context.Context, code string) *AuthorizationCode {
	authCode, err := sqlx.OneContext[AuthorizationCode](
		ctx,
		db.DB(),
		`SELECT * FROM authorization_codes WHERE code = $1 LIMIT 1`,
		code,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get authorization code by code", "err", err)
		return nil
	}

	if authCode.Code, err = cryptography.DecryptAES(authCode.Code, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt auth code", "err", err)
		return nil
	}

	return &authCode
}

func (db *Database) CreateAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	encryptedCode, err := cryptography.EncryptAES(code.Code, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = db.DB().ExecContext(
		ctx,
		`
		INSERT INTO authorization_codes (
			id, user_id, application, code, redirect_uri, scopes, state, nonce,
			code_challenge, code_challenge_method, created_at, expires_at, redeemed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`,
		code.ID,
		code.UserID,
		code.Application,
		encryptedCode,
		code.RedirectURI,
		code.Scopes,
		code.State,
		code.Nonce,
		code.CodeChallenge,
		code.CodeChallengeMethod,
		code.CreatedAt,
		code.ExpiresAt,
		code.RedeemedAt,
	)

	return err
}

func (db *Database) DeleteAuthorizationCodeByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM authorization_codes WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) DeleteDeadAuthCodes(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM authorization_codes WHERE expires_at < $1 OR redeemed_at < $1`,
		time.Now().Add(-time.Hour),
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) MarkAuthorizationCodeAsRedeemed(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE authorization_codes SET redeemed_at = $1 WHERE id = $2`,
		time.Now(),
		id,
	)

	return err
}
