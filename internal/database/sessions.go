package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const SessionKey = "sess"
const SessionFamilyKey = "sfam"

type SessionState string

const (
	SessionStateActive  SessionState = "active"
	SessionStateRotated SessionState = "rotated"
)

type Session struct {
	ID              string         `json:"id"`
	FamilyID        string         `json:"family_id" db:"family_id"`
	UserID          string         `json:"user_id" db:"user_id"`
	AccountID       *string        `json:"account_id" db:"account_id"`
	ApplicationID   string         `json:"application_id" db:"application_id"`
	Address         string         `json:"ip_address" db:"ip_address"`
	UserAgent       string         `json:"user_agent" db:"user_agent"`
	Device          string         `json:"device"`
	Geolocation     *string        `json:"geolocation"`
	Token           string         `json:"token"`
	Scopes          pq.StringArray `json:"scopes"`
	Status          SessionState   `json:"status"`
	AuthenticatedAt time.Time      `json:"authenticated_at" db:"authenticated_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	ExpiresAt       time.Time      `json:"expires_at" db:"expires_at"`
}

func (db *Database) GetSessionByFamily(ctx context.Context, familyID string) *Session {
	session, err := sqlx.OneContext[Session](
		ctx,
		db.DB(),
		`SELECT * FROM sessions WHERE family_id = $1 AND status = $2 LIMIT 1`,
		familyID,
		SessionStateActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get session by family", "err", err)
		return nil
	}

	if session.Address, err = cryptography.DecryptAES(session.Address, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt session ip address", "err", err)
		return nil
	}

	return &session
}

func (db *Database) GetUserSessions(ctx context.Context, userId string) []Session {
	sessions, err := list[Session](
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

	if err := db.decryptSessionAddresses(sessions); err != nil {
		slog.Error("failed to decrypt session ip address", "err", err)
		return []Session{}
	}

	return sessions
}

func (db *Database) GetActiveUserSessions(ctx context.Context, userID string, before *Cursor, limit int) ([]Session, error) {
	query := `SELECT * FROM sessions WHERE user_id = $1 AND status = $2 AND expires_at > $3`
	args := []any{userID, SessionStateActive, time.Now(), limit}

	if before != nil {
		query += ` AND (created_at, id) < ($5, $6)`
		args = append(args, before.CreatedAt, before.ID)
	}

	query += ` ORDER BY created_at DESC, id DESC LIMIT $4`

	sessions, err := list[Session](ctx, db.DB(), query, args...)
	if err != nil {
		return nil, err
	}

	if err := db.decryptSessionAddresses(sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (db *Database) decryptSessionAddresses(sessions []Session) error {
	for idx, session := range sessions {
		address, err := cryptography.DecryptAES(session.Address, db.config.AppToken)
		if err != nil {
			return err
		}

		sessions[idx].Address = address
	}

	return nil
}

func (db *Database) GetSessionByToken(ctx context.Context, token string) *Session {
	hashed := cryptography.HMAC(token, db.config.AppToken)

	session, err := sqlx.OneContext[Session](
		ctx,
		db.DB(),
		`SELECT * FROM sessions WHERE token = $1 LIMIT 1`,
		hashed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get session by token", "err", err)
		return nil
	}

	if session.Address, err = cryptography.DecryptAES(session.Address, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt session ip address", "err", err)
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

func (db *Database) DeleteDeadSessions(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM sessions WHERE expires_at < $1 OR (status != $2 AND updated_at < $1)`,
		time.Now().Add(-time.Hour*6),
		SessionStateActive,
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) CreateSession(ctx context.Context, session *Session) error {
	return db.insertSession(ctx, db.DB(), session)
}

func (db *Database) insertSession(ctx context.Context, execer sqlx.ExecerContext, session *Session) error {
	hashedToken := cryptography.HMAC(session.Token, db.config.AppToken)

	encryptedAddress, err := cryptography.EncryptAES(session.Address, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = execer.ExecContext(
		ctx,
		`
		INSERT INTO sessions (
			id, family_id, user_id, account_id, application_id,
			ip_address, user_agent, geolocation, device,
			token, scopes, status, authenticated_at, created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`,
		session.ID,
		session.FamilyID,
		session.UserID,
		session.AccountID,
		session.ApplicationID,
		encryptedAddress,
		session.UserAgent,
		session.Geolocation,
		session.Device,
		hashedToken,
		session.Scopes,
		session.Status,
		session.AuthenticatedAt,
		session.CreatedAt,
		time.Now(),
		session.ExpiresAt,
	)

	return err
}

func (db *Database) RotateSession(ctx context.Context, current *Session, next *Session) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE sessions
		SET status = $2, updated_at = $3
		WHERE id = $1 AND status = $4
		`,
		current.ID,
		SessionStateRotated,
		time.Now(),
		SessionStateActive,
	)

	if err != nil {
		return false, err
	}

	rotated, err := result.RowsAffected()
	if err != nil || rotated == 0 {
		return false, err
	}

	if err := db.insertSession(ctx, tx, next); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) GetSessionCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COUNT(*) FROM sessions WHERE status = $1 AND expires_at > $2`,
		SessionStateActive,
		time.Now(),
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
