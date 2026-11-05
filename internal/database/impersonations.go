package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

const ImpersonationKey = "imp"

type Impersonation struct {
	ID             string    `json:"id"`
	ImpersonatorID string    `json:"impersonator_id" db:"impersonator_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	ApplicationID  string    `json:"application_id" db:"application_id"`
	FamilyID       string    `json:"family_id" db:"family_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
}

type SharedApplication struct {
	Application
	Authorized bool `json:"authorized"`
}

func (db *Database) SaveImpersonation(ctx context.Context, impersonation *Impersonation) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO impersonations (
			id, impersonator_id, user_id, application_id, family_id, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (impersonator_id, application_id) DO UPDATE SET
			id = EXCLUDED.id,
			user_id = EXCLUDED.user_id,
			family_id = EXCLUDED.family_id,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at
		`,
		impersonation.ID,
		impersonation.ImpersonatorID,
		impersonation.UserID,
		impersonation.ApplicationID,
		impersonation.FamilyID,
		impersonation.CreatedAt,
		impersonation.ExpiresAt,
	)

	return err
}

func (db *Database) GetImpersonation(ctx context.Context, impersonatorID string, familyID string, applicationID string) *Impersonation {
	impersonation, err := sqlx.OneContext[Impersonation](
		ctx,
		db.DB(),
		`
		SELECT * FROM impersonations
		WHERE impersonator_id = $1 AND family_id = $2 AND application_id = $3 AND expires_at > NOW()
		LIMIT 1
		`,
		impersonatorID,
		familyID,
		applicationID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get impersonation", "err", err)
		return nil
	}

	return &impersonation
}

func (db *Database) ClaimImpersonation(ctx context.Context, id string) (bool, error) {
	result, err := db.DB().ExecContext(ctx, `DELETE FROM impersonations WHERE id = $1`, id)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (db *Database) DeleteDeadImpersonations(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(ctx, `DELETE FROM impersonations WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) GetSharedApplications(
	ctx context.Context,
	impersonatorID string,
	userID string,
	query string,
	page int,
	limit int,
) ([]SharedApplication, error) {
	return list[SharedApplication](
		ctx,
		db.DB(),
		`
		SELECT applications.*, EXISTS (
			SELECT 1 FROM application_grants
			WHERE application_grants.application_id = applications.id AND application_grants.user_id = $2
		) AS authorized
		FROM applications
		WHERE NOT applications.internal
		AND applications.name ILIKE '%' || $3 || '%'
		AND (
			NOT applications.role_restricted
			OR (
				EXISTS (
					SELECT 1 FROM application_role_bindings
					INNER JOIN role_bindings ON application_role_bindings.role_id = role_bindings.role_id
					WHERE application_role_bindings.application_id = applications.id AND role_bindings.user_id = $1
				)
				AND EXISTS (
					SELECT 1 FROM application_role_bindings
					INNER JOIN role_bindings ON application_role_bindings.role_id = role_bindings.role_id
					WHERE application_role_bindings.application_id = applications.id AND role_bindings.user_id = $2
				)
			)
		)
		ORDER BY authorized DESC, applications.name, applications.id
		LIMIT $4 OFFSET $5
		`,
		impersonatorID,
		userID,
		query,
		limit,
		(page-1)*limit,
	)
}
