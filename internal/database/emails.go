package database

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const EmailKey = "email"

type UserEmail struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id" db:"user_id"`
	Email         string     `json:"email"`
	Verified      bool       `json:"verified"`
	SelfAsserted  bool       `json:"-" db:"self_asserted"`
	Code          string     `json:"-"`
	CodeAttempts  int        `json:"-" db:"code_attempts"`
	CodeExpiresAt *time.Time `json:"-" db:"code_expires_at"`
	Primary       bool       `json:"primary" db:"preferred"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

func (e *UserEmail) Trusted() bool {
	return e.Verified && !e.SelfAsserted
}

func (db *Database) GetEmailsByUserID(ctx context.Context, userID string) []UserEmail {
	emails, err := list[UserEmail](
		ctx,
		db.DB(),
		`SELECT * FROM user_emails WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []UserEmail{}
	} else if err != nil {
		slog.Error("failed to get emails by user id", "err", err)
		return []UserEmail{}
	}

	return emails
}

func (db *Database) GetEmailByID(ctx context.Context, id string) *UserEmail {
	userEmail, err := sqlx.OneContext[UserEmail](
		ctx,
		db.DB(),
		`SELECT * FROM user_emails WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get email by id", "err", err)
		return nil
	}

	return &userEmail
}

func (db *Database) GetEmailByName(ctx context.Context, email string) *UserEmail {
	userEmail, err := sqlx.OneContext[UserEmail](
		ctx,
		db.DB(),
		`SELECT * FROM user_emails WHERE LOWER(email) = LOWER($1) ORDER BY verified DESC, created_at DESC LIMIT 1`,
		email,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get email by name", "err", err)
		return nil
	}

	return &userEmail
}

func (db *Database) CreateEmail(ctx context.Context, email *UserEmail) error {
	return duplicate(db.insertEmail(ctx, db.DB(), email))
}

func (db *Database) insertEmail(ctx context.Context, execer sqlx.ExecerContext, email *UserEmail) error {
	_, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO user_emails (id, user_id, email, created_at, code, code_expires_at, verified, preferred, self_asserted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		email.ID,
		email.UserID,
		email.Email,
		email.CreatedAt,
		cryptography.HMAC(email.Code, db.config.AppToken),
		time.Now().Add(time.Minute*15),
		email.Verified,
		email.Primary,
		email.SelfAsserted,
	)

	if err != nil || !email.Verified {
		return err
	}

	return db.deletePendingEmails(ctx, execer, email.ID, email.Email)
}

func (db *Database) deletePendingEmails(ctx context.Context, execer sqlx.ExecerContext, verifiedID string, email string) error {
	_, err := execer.ExecContext(
		ctx,
		`DELETE FROM user_emails WHERE LOWER(email) = LOWER($1) AND verified = FALSE AND id != $2`,
		email,
		verifiedID,
	)

	return err
}

func (db *Database) ResetEmailCode(ctx context.Context, id string, code string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE user_emails
		SET code = $2, code_expires_at = $3, code_attempts = 0
		WHERE id = $1 AND verified = FALSE AND (code_expires_at IS NULL OR code_expires_at < $4)
		`,
		id,
		cryptography.HMAC(code, db.config.AppToken),
		time.Now().Add(time.Minute*15),
		time.Now().Add(time.Minute*14),
	)

	if err != nil {
		return false, err
	}

	updated, err := result.RowsAffected()
	return updated > 0, err
}

var ErrEmailOwnerRequired = errors.New("the email must be verified while signed in as the account it was added to")

func (db *Database) VerifyEmailCode(ctx context.Context, email string, code string, requesterID *string) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	candidates, err := list[UserEmail](
		ctx,
		tx,
		`
		UPDATE user_emails
		SET code_attempts = code_attempts + 1
		WHERE LOWER(email) = LOWER($1) AND verified = FALSE AND code_expires_at > $2 AND code_attempts < 5
		RETURNING *
		`,
		email,
		time.Now(),
	)

	if err != nil {
		return false, err
	}

	hashed := cryptography.HMAC(code, db.config.AppToken)
	index := slices.IndexFunc(candidates, func(candidate UserEmail) bool {
		return subtle.ConstantTimeCompare([]byte(candidate.Code), []byte(hashed)) == 1
	})

	if index == -1 {
		return false, tx.Commit()
	}

	candidate := candidates[index]
	if requesterID == nil || *requesterID != candidate.UserID {
		established, err := sqlx.OneContext[bool](
			ctx,
			tx,
			`SELECT EXISTS (SELECT 1 FROM user_emails WHERE user_id = $1 AND id <> $2 AND verified = TRUE)`,
			candidate.UserID,
			candidate.ID,
		)

		if err != nil {
			return false, err
		}

		if established {
			if err := tx.Commit(); err != nil {
				return false, err
			}

			return false, ErrEmailOwnerRequired
		}
	}

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE user_emails
		SET verified = TRUE, code_expires_at = NULL
		WHERE id = $1 AND NOT EXISTS (
			SELECT 1 FROM user_emails WHERE LOWER(email) = LOWER($2) AND verified = TRUE
		)
		`,
		candidate.ID,
		email,
	)

	if err != nil {
		return false, err
	}

	verified, err := result.RowsAffected()
	if err != nil {
		return false, err
	} else if verified == 0 {
		return false, tx.Commit()
	}

	if err := db.deletePendingEmails(ctx, tx, candidate.ID, email); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) DeleteEmail(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM user_emails WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) GetPrimaryEmail(ctx context.Context, userID string) *UserEmail {
	email, err := sqlx.OneContext[UserEmail](
		ctx,
		db.DB(),
		`SELECT * FROM user_emails WHERE user_id = $1 AND preferred = TRUE LIMIT 1`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get preferred email", "err", err)
		return nil
	}

	return &email
}

func (db *Database) SetPrimaryEmail(ctx context.Context, userID string, emailID string) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		`UPDATE user_emails SET preferred = FALSE WHERE user_id = $1 AND preferred = TRUE`,
		userID,
	)

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE user_emails SET preferred = TRUE WHERE id = $1`,
		emailID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) EmailClaimed(ctx context.Context, email string) bool {
	exists, err := sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`
		SELECT EXISTS(
			SELECT 1 FROM user_emails
			WHERE LOWER(email) = LOWER($1) AND (verified = TRUE OR created_at > $2)
		)
		`,
		email,
		time.Now().Add(-time.Hour),
	)

	if err != nil {
		slog.Error("failed to check if email is claimed", "err", err)
		return true
	}

	return exists
}
