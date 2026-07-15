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

const TotpKey = "totp"

type Totp struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Secret    string    `json:"secret"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (db *Database) GetTotpByID(ctx context.Context, id string) *Totp {
	totp, err := sqlx.OneContext[Totp](
		ctx,
		db.DB(),
		`SELECT * FROM totp WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get totp by id", "err", err)
		return nil
	}

	if totp.Secret, err = cryptography.DecryptAES(totp.Secret, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt totp secret", "err", err)
		return nil
	}

	return &totp
}

func (db *Database) GetUserTotp(ctx context.Context, userId string) *Totp {
	totp, err := sqlx.OneContext[Totp](
		ctx,
		db.DB(),
		`SELECT * FROM totp WHERE user_id = $1 LIMIT 1`,
		userId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get totp by user", "err", err)
		return nil
	}

	if totp.Secret, err = cryptography.DecryptAES(totp.Secret, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt totp secret", "err", err)
		return nil
	}

	return &totp
}

func (db *Database) DeleteTotpByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM totp WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CreateTotp(ctx context.Context, totp *Totp) error {
	encrypted, err := cryptography.EncryptAES(totp.Secret, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = db.DB().ExecContext(
		ctx,
		`
		INSERT INTO totp (
			id, user_id, secret, verified, created_at
		) VALUES ($1, $2, $3, $4, $5)
		`,
		totp.ID,
		totp.UserID,
		encrypted,
		totp.Verified,
		totp.CreatedAt,
	)

	return err
}

func (db *Database) UpdateTotp(ctx context.Context, totp *Totp) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE totp
		SET verified = $2
		WHERE id = $1
		`,
		totp.ID,
		totp.Verified,
	)

	return err
}

func (db *Database) IsTotpEnabled(ctx context.Context, userId string) (bool, error) {
	count, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM totp WHERE user_id = $1 AND verified = true LIMIT 1`,
		userId,
	)

	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func (db *Database) GetTotpCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM totp`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
