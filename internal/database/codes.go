package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const AuthCodeKey = "code"

type AuthorizationCode struct {
	ID string `json:"id"`

	UserID        string `json:"user_id" db:"user_id"`
	ApplicationID string `json:"application_id" db:"application_id"`

	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri" db:"redirect_uri"`

	Scopes pq.StringArray `json:"scopes"`
	State  *string        `json:"state"`
	Nonce  *string        `json:"nonce"`

	CodeChallenge       *string `json:"code_challenge" db:"code_challenge"`
	CodeChallengeMethod *string `json:"code_challenge_method" db:"code_challenge_method"`

	FamilyID *string `json:"family_id" db:"family_id"`

	AuthenticatedAt time.Time `json:"authenticated_at" db:"authenticated_at"`

	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	RedeemedAt *time.Time `json:"redeemed_at" db:"redeemed_at"`
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

	return &authCode
}

func (db *Database) GetAuthorizationCodeByCode(ctx context.Context, code string) *AuthorizationCode {
	hashed := cryptography.HMAC(code, db.config.AppToken)

	authCode, err := sqlx.OneContext[AuthorizationCode](
		ctx,
		db.DB(),
		`SELECT * FROM authorization_codes WHERE code = $1 LIMIT 1`,
		hashed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get authorization code by code", "err", err)
		return nil
	}

	return &authCode
}

func (db *Database) CreateAuthorizationCode(ctx context.Context, code *AuthorizationCode) error {
	hashedCode := cryptography.HMAC(code.Code, db.config.AppToken)

	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO authorization_codes (
			id, user_id, application_id, code, redirect_uri, scopes, state, nonce,
			code_challenge, code_challenge_method, authenticated_at, created_at, expires_at, redeemed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
		code.ID,
		code.UserID,
		code.ApplicationID,
		hashedCode,
		code.RedirectURI,
		code.Scopes,
		code.State,
		code.Nonce,
		code.CodeChallenge,
		code.CodeChallengeMethod,
		code.AuthenticatedAt,
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

func (db *Database) RedeemAuthorizationCode(ctx context.Context, id string, familyID string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`UPDATE authorization_codes SET redeemed_at = $1, family_id = $2 WHERE id = $3 AND redeemed_at IS NULL`,
		time.Now(),
		familyID,
		id,
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (db *Database) GetAuthorizationCodesByUserID(ctx context.Context, userID string) []AuthorizationCode {
	authCodes, err := list[AuthorizationCode](
		ctx,
		db.DB(),
		`SELECT * FROM authorization_codes WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if err != nil {
		slog.Error("failed to get user authorization codes", "err", err)
		return []AuthorizationCode{}
	}

	return authCodes
}
