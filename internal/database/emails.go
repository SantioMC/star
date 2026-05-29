package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

const EmailKey = "email"

type UserEmail struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	Code      string    `json:"code"`
	Primary   bool      `json:"primary" db:"preferred"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (db *Database) GetEmailsByUserID(ctx context.Context, userID string) []UserEmail {
	emails, err := sqlx.ListContext[UserEmail](
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
		`SELECT * FROM user_emails WHERE email = $1 LIMIT 1`,
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
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO user_emails (id, user_id, email, created_at, code, verified, preferred)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		email.ID,
		email.UserID,
		email.Email,
		email.CreatedAt,
		email.Code,
		email.Verified,
		email.Primary,
	)

	return err
}

func (db *Database) UpdateEmail(ctx context.Context, email *UserEmail) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE user_emails
		SET verified = $2 AND preferred = $3 AND code = $4
		WHERE id = $1
		`,
		email.ID,
		email.Verified,
		email.Primary,
		email.Code,
	)

	return err
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
	tx, err := db.DB().Beginx()
	if err != nil {
		return err
	}

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

func (db *Database) EmailExists(ctx context.Context, email string) bool {
	exists, err := sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`SELECT EXISTS(SELECT 1 FROM user_emails WHERE email = $1)`,
		email,
	)

	if err != nil {
		slog.Error("failed to check if email exists", "err", err)
		return false
	}

	return exists
}
