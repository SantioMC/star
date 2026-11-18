package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vinovest/sqlx"
)

var ErrLoginLocked = errors.New("the user has too many failed login attempts")

type LoginAttempts struct {
	UserID          string     `json:"user_id" db:"user_id"`
	Attempts        int        `json:"attempts"`
	LastAttemptAt   time.Time  `json:"last_attempt_at" db:"last_attempt_at"`
	WindowStartedAt time.Time  `json:"window_started_at" db:"window_started_at"`
	LockedUntil     *time.Time `json:"locked_until" db:"locked_until"`
}

type LockoutPolicy struct {
	Threshold int
	Window    time.Duration
	Duration  time.Duration
}

func (attempts *LoginAttempts) IsLocked() bool {
	return attempts.LockedUntil != nil && attempts.LockedUntil.After(time.Now())
}

func (db *Database) GetLoginAttempts(ctx context.Context, userID string) (*LoginAttempts, error) {
	attempts, err := sqlx.OneContext[LoginAttempts](
		ctx,
		db.DB(),
		`SELECT * FROM login_attempts WHERE user_id = $1`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &attempts, nil
}

func (db *Database) ReserveLoginAttempt(ctx context.Context, userID string, policy LockoutPolicy) (*LoginAttempts, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	now := time.Now()

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO login_attempts (user_id, attempts, last_attempt_at, window_started_at)
		VALUES ($1, 0, $2, $2)
		ON CONFLICT (user_id) DO NOTHING
		`,
		userID,
		now,
	)

	if err != nil {
		return nil, err
	}

	current, err := sqlx.OneContext[LoginAttempts](
		ctx,
		tx,
		`SELECT * FROM login_attempts WHERE user_id = $1 FOR UPDATE`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	if current.IsLocked() {
		return nil, ErrLoginLocked
	}

	next := LoginAttempts{
		UserID:          userID,
		Attempts:        current.Attempts + 1,
		LastAttemptAt:   now,
		WindowStartedAt: current.WindowStartedAt,
	}

	lockExpired := current.LockedUntil != nil
	windowExpired := now.Sub(current.WindowStartedAt) > policy.Window

	if lockExpired || windowExpired {
		next.Attempts = 1
		next.WindowStartedAt = now
	}

	if next.Attempts >= policy.Threshold {
		next.LockedUntil = new(now.Add(policy.Duration))
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE login_attempts
		SET attempts = $2, last_attempt_at = $3, window_started_at = $4, locked_until = $5
		WHERE user_id = $1
		`,
		userID,
		next.Attempts,
		next.LastAttemptAt,
		next.WindowStartedAt,
		next.LockedUntil,
	)

	if err != nil {
		return nil, err
	}

	return &next, tx.Commit()
}

func (db *Database) ReleaseLoginAttempt(ctx context.Context, userID string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE login_attempts
		SET attempts = GREATEST(attempts - 1, 0), locked_until = NULL
		WHERE user_id = $1
		`,
		userID,
	)

	return err
}

func (db *Database) ClearLoginAttempts(ctx context.Context, userID string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM login_attempts WHERE user_id = $1`,
		userID,
	)

	return err
}
