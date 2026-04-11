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

type Account struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	Password    *string                `json:"password"`
	PasswordAlg cryptography.Algorithm `json:"password_alg" db:"password_alg"`

	UserID     string    `json:"user_id" db:"user_id"`
	ExternalID *string   `json:"external_id" db:"external_id"`
	Name       *string   `json:"name" db:"account_name"`
	Email      *string   `json:"email" db:"account_email"`
	Avatar     *string   `json:"avatar" db:"account_avatar"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (db *Database) GetAccountsByUserID(ctx context.Context, userID string) []Account {
	accounts, err := sqlx.ListContext[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get accounts by user id", "err", err)
		return nil
	}

	return accounts
}

func (db *Database) GetInternalAccountByUserID(ctx context.Context, userID string) *Account {
	account, err := sqlx.OneContext[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE user_id = $1 AND type = $2 LIMIT 1`,
		userID,
		"internal",
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get internal account by user id", "err", err)
		return nil
	}

	return &account
}

func (db *Database) CreateAccount(ctx context.Context, account *Account) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO accounts (id, user_id, type, password, password_alg, external_id, account_name, account_email, account_avatar, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
		account.ID,
		account.UserID,
		account.Type,
		account.Password,
		account.PasswordAlg,
		account.ExternalID,
		account.Name,
		account.Email,
		account.Avatar,
		account.CreatedAt,
	)

	return err
}

func (db *Database) UpdateAccount(ctx context.Context, account *Account) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE accounts
		SET password = $2, password_alg = $3, account_name = $4, account_email = $5, account_avatar = $6
		WHERE id = $1
		`,
		account.ID,
		account.Password,
		account.PasswordAlg,
		account.Name,
		account.Email,
		account.Avatar,
	)

	return err
}

func (db *Database) DeleteAccountByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM accounts WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) GetAccountsCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM accounts`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
