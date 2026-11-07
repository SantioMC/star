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

const PasswordResetKey = "pwr"

type PasswordReset struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func CreatePasswordResetToken() string {
	return cryptography.CreateSecureString(48)
}

func (db *Database) GetPasswordResetByToken(ctx context.Context, token string) *PasswordReset {
	hashed := cryptography.HMAC(token, db.config.AppToken)

	reset, err := sqlx.OneContext[PasswordReset](
		ctx,
		db.DB(),
		`SELECT * FROM password_resets WHERE token = $1 AND expires_at > $2 LIMIT 1`,
		hashed,
		time.Now(),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get password reset by token", "err", err)
		return nil
	}

	return &reset
}

func (db *Database) IssuePasswordReset(ctx context.Context, reset *PasswordReset) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET id = EXCLUDED.id, token = EXCLUDED.token, expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at
		WHERE password_resets.created_at < $6
		`,
		reset.ID,
		reset.UserID,
		cryptography.HMAC(reset.Token, db.config.AppToken),
		reset.ExpiresAt,
		reset.CreatedAt,
		reset.CreatedAt.Add(-time.Minute),
	)

	if err != nil {
		return false, err
	}

	issued, err := result.RowsAffected()
	return issued > 0, err
}

func (db *Database) CompletePasswordReset(ctx context.Context, reset *PasswordReset, hashed string, alg cryptography.Algorithm) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	consumed, err := tx.ExecContext(
		ctx,
		`DELETE FROM password_resets WHERE id = $1 AND expires_at > $2`,
		reset.ID,
		time.Now(),
	)

	if err != nil {
		return false, err
	}

	if affected, err := consumed.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	updated, err := tx.ExecContext(
		ctx,
		`UPDATE accounts SET password = $2, password_alg = $3, password_change_required = FALSE WHERE user_id = $1 AND type = 'internal' AND password IS NOT NULL`,
		reset.UserID,
		hashed,
		alg,
	)

	if err != nil {
		return false, err
	}

	if affected, err := updated.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM login_attempts WHERE user_id = $1`, reset.UserID); err != nil {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, reset.UserID); err != nil {
		return false, err
	}

	if err := db.revokeAuthorizations(ctx, tx, &reset.UserID); err != nil {
		return false, err
	}

	if err := db.revokeUser(ctx, tx, reset.UserID, RevocationReasonPasswordReset); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) DeleteExpiredPasswordResets(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM password_resets WHERE expires_at < $1`,
		time.Now(),
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
