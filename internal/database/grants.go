package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/vinovest/sqlx"
)

const ApplicationGrantKey = "grant"

type ApplicationGrant struct {
	ID            string         `json:"id"`
	UserID        string         `json:"user_id" db:"user_id"`
	ApplicationID string         `json:"application_id" db:"application_id"`
	Scopes        pq.StringArray `json:"scopes"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

type UserApplicationGrant struct {
	ApplicationGrant
	ApplicationName string     `json:"application_name" db:"application_name"`
	ApplicationURL  string     `json:"application_url" db:"application_url"`
	LastUsedAt      *time.Time `json:"last_used_at" db:"last_used_at"`
}

func (db *Database) GetApplicationGrant(ctx context.Context, userID string, applicationID string) *ApplicationGrant {
	grant, err := sqlx.OneContext[ApplicationGrant](
		ctx,
		db.DB(),
		`SELECT * FROM application_grants WHERE user_id = $1 AND application_id = $2 LIMIT 1`,
		userID,
		applicationID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application grant", "err", err)
		return nil
	}

	return &grant
}

func (db *Database) GetUserApplicationGrants(ctx context.Context, userID string) []UserApplicationGrant {
	grants, err := list[UserApplicationGrant](
		ctx,
		db.DB(),
		`
		SELECT
			grant_entry.*,
			app.name AS application_name,
			app.application_url AS application_url,
			(
				SELECT MAX(session.updated_at) FROM sessions session
				WHERE session.user_id = grant_entry.user_id AND session.application_id = grant_entry.application_id
			) AS last_used_at
		FROM application_grants grant_entry
		JOIN applications app ON app.id = grant_entry.application_id
		WHERE grant_entry.user_id = $1
		ORDER BY grant_entry.updated_at DESC
		`,
		userID,
	)

	if err != nil {
		slog.Error("failed to get user application grants", "err", err)
		return []UserApplicationGrant{}
	}

	return grants
}

func (db *Database) SaveApplicationGrant(ctx context.Context, grant *ApplicationGrant) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO application_grants (id, user_id, application_id, scopes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, application_id)
		DO UPDATE SET scopes = EXCLUDED.scopes, updated_at = EXCLUDED.updated_at
		`,
		grant.ID,
		grant.UserID,
		grant.ApplicationID,
		grant.Scopes,
		grant.CreatedAt,
		grant.UpdatedAt,
	)

	return err
}

func (db *Database) RevokeApplicationGrant(ctx context.Context, userID string, applicationID string) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	statements := []string{
		`DELETE FROM application_grants WHERE user_id = $1 AND application_id = $2`,
		`DELETE FROM authorization_codes WHERE user_id = $1 AND application_id = $2`,
	}

	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, userID, applicationID); err != nil {
			return err
		}
	}

	err = db.revoke(
		ctx,
		tx,
		RevocationKindSession,
		RevocationReasonGrantRevoked,
		`DELETE FROM sessions WHERE user_id = $1 AND application_id = $2 RETURNING family_id AS subject, user_id`,
		userID,
		applicationID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}
