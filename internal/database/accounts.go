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

const AccountKey = "acc"

type Account struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	Password               *string                `json:"password"`
	PasswordAlg            cryptography.Algorithm `json:"password_alg" db:"password_alg"`
	PasswordChangeRequired bool                   `json:"password_change_required" db:"password_change_required"`

	UserID     string    `json:"user_id" db:"user_id"`
	ProviderID *string   `json:"provider_id" db:"provider_id"`
	ExternalID *string   `json:"external_id" db:"external_id"`
	Name       *string   `json:"name" db:"account_name"`
	Email      *string   `json:"email" db:"account_email"`
	Avatar     *string   `json:"avatar" db:"account_avatar"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (db *Database) GetAccountsByUserID(ctx context.Context, userID string) []Account {
	accounts, err := list[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Account{}
	} else if err != nil {
		slog.Error("failed to get accounts by user id", "err", err)
		return []Account{}
	}

	return accounts
}

func (db *Database) GetAccountByID(ctx context.Context, id string) *Account {
	account, err := sqlx.OneContext[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get account by id", "err", err)
		return nil
	}

	return &account
}

func (db *Database) GetProviderAccount(ctx context.Context, providerID string, externalID string) *Account {
	account, err := sqlx.OneContext[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE provider_id = $1 AND external_id = $2 LIMIT 1`,
		providerID,
		externalID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get provider account", "err", err)
		return nil
	}

	return &account
}

func (db *Database) GetUserProviderAccount(ctx context.Context, userID string, providerID string) *Account {
	account, err := sqlx.OneContext[Account](
		ctx,
		db.DB(),
		`SELECT * FROM accounts WHERE user_id = $1 AND provider_id = $2 LIMIT 1`,
		userID,
		providerID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get user provider account", "err", err)
		return nil
	}

	return &account
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
	return duplicate(db.insertAccount(ctx, db.DB(), account))
}

func (db *Database) insertAccount(ctx context.Context, execer sqlx.ExecerContext, account *Account) error {
	_, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO accounts (id, user_id, type, password, password_alg, password_change_required, provider_id, external_id, account_name, account_email, account_avatar, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
		account.ID,
		account.UserID,
		account.Type,
		account.Password,
		account.PasswordAlg,
		account.PasswordChangeRequired,
		account.ProviderID,
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
		SET password = $2, password_alg = $3, account_name = $4, account_email = $5, account_avatar = $6, password_change_required = $7
		WHERE id = $1
		`,
		account.ID,
		account.Password,
		account.PasswordAlg,
		account.Name,
		account.Email,
		account.Avatar,
		account.PasswordChangeRequired,
	)

	return err
}

func (db *Database) RequirePasswordChange(ctx context.Context, userID string) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	updated, err := tx.ExecContext(
		ctx,
		`UPDATE accounts SET password_change_required = TRUE WHERE user_id = $1 AND type = 'internal' AND password IS NOT NULL`,
		userID,
	)

	if err != nil {
		return false, err
	}

	if affected, err := updated.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return false, err
	}

	if err := db.revokeUser(ctx, tx, userID, RevocationReasonPasswordChangeRequired); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) RequirePasswordChangeForAll(ctx context.Context) (int64, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	updated, err := tx.ExecContext(
		ctx,
		`UPDATE accounts SET password_change_required = TRUE WHERE type = 'internal' AND password IS NOT NULL`,
	)

	if err != nil {
		return 0, err
	}

	affected, err := updated.RowsAffected()
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions`); err != nil {
		return 0, err
	}

	if err := db.revokeAll(ctx, tx, RevocationReasonPasswordChangeRequired); err != nil {
		return 0, err
	}

	return affected, tx.Commit()
}

func (db *Database) DeleteAccountByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM accounts WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CanRemoveAccount(ctx context.Context, account *Account) (bool, error) {
	return canRemoveAccount(ctx, db.DB(), account)
}

func canRemoveAccount(ctx context.Context, queryer sqlx.QueryerContext, account *Account) (bool, error) {
	return sqlx.OneContext[bool](
		ctx,
		queryer,
		`
		SELECT EXISTS (
			SELECT 1 FROM accounts
			LEFT JOIN providers ON providers.id = accounts.provider_id
			WHERE accounts.user_id = $1
			AND accounts.id != $2
			AND (accounts.password IS NOT NULL OR providers.enabled)
		) OR EXISTS (
			SELECT 1 FROM passkeys WHERE user_id = $1
		)
		`,
		account.UserID,
		account.ID,
	)
}

func (db *Database) RemoveProviderAccount(ctx context.Context, account *Account, familyID string) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, account.UserID); err != nil {
		return false, err
	}

	if removable, err := canRemoveAccount(ctx, tx, account); err != nil || !removable {
		return false, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE sessions SET account_id = NULL WHERE account_id = $1 AND family_id = $2`,
		account.ID,
		familyID,
	)

	if err != nil {
		return false, err
	}

	err = db.revoke(
		ctx,
		tx,
		RevocationKindSession,
		RevocationReasonConnectionRemoved,
		`SELECT family_id AS subject, user_id FROM sessions WHERE account_id = $1`,
		account.ID,
	)

	if err != nil {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, account.ID); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) GetAccountsCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM accounts`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
