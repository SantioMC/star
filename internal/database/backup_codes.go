package database

import (
	"context"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const BackupCodeKey = "bkc"

func (db *Database) hashBackupCode(code string) string {
	return cryptography.HMAC(code, db.config.AppToken)
}

func (db *Database) ReplaceBackupCodes(ctx context.Context, userID string, codes []string) error {
	tx, err := db.DB().Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}

	for _, code := range codes {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO backup_codes (id, user_id, code, created_at) VALUES ($1, $2, $3, $4)`,
			CreateID(BackupCodeKey),
			userID,
			db.hashBackupCode(code),
			time.Now(),
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Database) ConsumeBackupCode(ctx context.Context, userID string, code string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM backup_codes WHERE user_id = $1 AND code = $2`,
		userID,
		db.hashBackupCode(code),
	)

	if err != nil {
		return false, err
	}

	deleted, err := result.RowsAffected()
	return deleted > 0, err
}

func (db *Database) CountBackupCodes(ctx context.Context, userID string) (int, error) {
	return sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM backup_codes WHERE user_id = $1`,
		userID,
	)
}

func (db *Database) DeleteBackupCodes(ctx context.Context, userID string) error {
	_, err := db.DB().ExecContext(ctx, `DELETE FROM backup_codes WHERE user_id = $1`, userID)
	return err
}
