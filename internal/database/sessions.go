package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

const SessionKey = "sess"

type SessionState string

const (
	SessionStateActive  SessionState = "active"
	SessionStateRotated SessionState = "rotated"
)

type Session struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id" db:"user_id"`
	AccountID     *string      `json:"account_id" db:"account_id"`
	ApplicationID string       `json:"application_id" db:"application_id"`
	Address       string       `json:"ip_address" db:"ip_address"`
	UserAgent     string       `json:"user_agent" db:"user_agent"`
	Device        string       `json:"device"`
	Geolocation   *string      `json:"geolocation"`
	Token         string       `json:"token"`
	Status        SessionState `json:"status"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	ExpiresAt     time.Time    `json:"expires_at" db:"expires_at"`
}

func (db *Database) GetSessionByID(ctx context.Context, id string) *Session {
	session, err := sqlx.OneContext[Session](
		ctx,
		db.DB(),
		`SELECT * FROM sessions WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get session by id", "err", err)
		return nil
	}

	return &session
}

func (db *Database) GetUserSessions(ctx context.Context, userId string) []Session {
	session, err := sqlx.ListContext[Session](
		ctx,
		db.DB(),
		`SELECT * FROM sessions WHERE user_id = $1`,
		userId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Session{}
	} else if err != nil {
		slog.Error("failed to get user sessions", "err", err)
		return []Session{}
	}

	return session
}

func (db *Database) GetSessionByToken(ctx context.Context, token string) *Session {
	session, err := sqlx.OneContext[Session](
		ctx,
		db.DB(),
		`SELECT * FROM sessions WHERE token = $1 LIMIT 1`,
		token,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get session by token", "err", err)
		return nil
	}

	return &session
}

func (db *Database) DeleteSessionByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM sessions WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) DeleteUserSessions(ctx context.Context, userId string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM sessions WHERE user_id = $1`,
		userId,
	)

	return err
}

func (db *Database) CreateSession(ctx context.Context, session *Session) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO sessions (
			id, user_id, account_id, application_id,
			ip_address, user_agent, geolocation, device,
			token, status, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
		session.ID,
		session.UserID,
		session.AccountID,
		session.ApplicationID,
		session.Address,
		session.UserAgent,
		session.Geolocation,
		session.Device,
		session.Token,
		session.Status,
		session.CreatedAt,
		session.ExpiresAt,
	)

	return err
}

func (db *Database) GetSessionCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM sessions`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
