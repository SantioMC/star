package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vinovest/sqlx"
)

const UserSuspensionKey = "sus"

type UserSuspension struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id" db:"user_id"`
	Reason          string     `json:"reason"`
	SuspendedBy     *string    `json:"suspended_by" db:"suspended_by"`
	SuspendedByName *string    `json:"suspended_by_name" db:"suspended_by_name"`
	ExpiresAt       *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	LiftedAt        *time.Time `json:"lifted_at" db:"lifted_at"`
	LiftedBy        *string    `json:"lifted_by" db:"lifted_by"`
	LiftedByName    *string    `json:"lifted_by_name" db:"lifted_by_name"`
	LiftReason      *string    `json:"lift_reason" db:"lift_reason"`
}

const selectUserSuspensions = `
	SELECT suspensions.*, suspender.name AS suspended_by_name, lifter.name AS lifted_by_name
	FROM user_suspensions suspensions
	LEFT JOIN users suspender ON suspender.id = suspensions.suspended_by
	LEFT JOIN users lifter ON lifter.id = suspensions.lifted_by
`

func (db *Database) GetActiveSuspension(ctx context.Context, userID string) (*UserSuspension, error) {
	suspension, err := sqlx.OneContext[UserSuspension](
		ctx,
		db.DB(),
		selectUserSuspensions+`
		WHERE suspensions.user_id = $1
			AND suspensions.lifted_at IS NULL
			AND (suspensions.expires_at IS NULL OR suspensions.expires_at > NOW())
		ORDER BY suspensions.created_at DESC
		LIMIT 1
		`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &suspension, nil
}

func (db *Database) GetUserSuspensions(ctx context.Context, userID string, before *Cursor, limit int) ([]UserSuspension, error) {
	query := selectUserSuspensions + ` WHERE suspensions.user_id = $1`
	args := []any{userID, limit}

	if before != nil {
		query += ` AND (suspensions.created_at, suspensions.id) < ($3, $4)`
		args = append(args, before.CreatedAt, before.ID)
	}

	query += ` ORDER BY suspensions.created_at DESC, suspensions.id DESC LIMIT $2`

	return list[UserSuspension](ctx, db.DB(), query, args...)
}

func (db *Database) SuspendUser(ctx context.Context, suspension *UserSuspension) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, suspension.UserID); err != nil {
		return err
	}

	result, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO user_suspensions (id, user_id, reason, suspended_by, expires_at, created_at)
		SELECT $1::VARCHAR, $2::VARCHAR, $3::VARCHAR, $4::VARCHAR, $5::TIMESTAMPTZ, $6::TIMESTAMPTZ
		WHERE NOT EXISTS (
			SELECT 1 FROM user_suspensions
			WHERE user_id = $2 AND lifted_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())
		)
		`,
		suspension.ID,
		suspension.UserID,
		suspension.Reason,
		suspension.SuspendedBy,
		suspension.ExpiresAt,
		suspension.CreatedAt,
	)

	if err != nil {
		return err
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	} else if inserted == 0 {
		return ErrDuplicate
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, suspension.UserID); err != nil {
		return err
	}

	if err := db.revokeUser(ctx, tx, suspension.UserID, RevocationReasonUserSuspended); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) LiftSuspension(ctx context.Context, userID string, liftedBy string, reason string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE user_suspensions
		SET lifted_at = NOW(), lifted_by = $2, lift_reason = $3
		WHERE user_id = $1 AND lifted_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())
		`,
		userID,
		liftedBy,
		reason,
	)

	if err != nil {
		return false, err
	}

	lifted, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return lifted > 0, nil
}
