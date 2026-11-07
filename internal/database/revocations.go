package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

type RevocationKind string

const (
	RevocationKindSession RevocationKind = "session"
	RevocationKindUser    RevocationKind = "user"
	RevocationKindGlobal  RevocationKind = "global"
)

type RevocationReason string

const (
	RevocationReasonSignedOut              RevocationReason = "signed_out"
	RevocationReasonSessionsCleared        RevocationReason = "sessions_cleared"
	RevocationReasonGrantRevoked           RevocationReason = "grant_revoked"
	RevocationReasonTokenReplayed          RevocationReason = "token_replayed"
	RevocationReasonCodeReplayed           RevocationReason = "code_replayed"
	RevocationReasonRolesChanged           RevocationReason = "roles_changed"
	RevocationReasonPasswordReset          RevocationReason = "password_reset"
	RevocationReasonPasswordChangeRequired RevocationReason = "password_change_required"
	RevocationReasonUserDeleted            RevocationReason = "user_deleted"
	RevocationReasonUserSuspended          RevocationReason = "user_suspended"
	RevocationReasonAppDeleted             RevocationReason = "application_deleted"
	RevocationReasonAppAccessChanged       RevocationReason = "application_access_changed"
	RevocationReasonConnectionRemoved      RevocationReason = "connection_removed"
	RevocationReasonManual                 RevocationReason = "manual"
)

const GlobalRevocationSubject = "*"

type TokenRevocation struct {
	Subject   string           `json:"subject"`
	Kind      RevocationKind   `json:"kind"`
	UserID    *string          `json:"user_id" db:"user_id"`
	Reason    RevocationReason `json:"reason"`
	RevokedAt time.Time        `json:"revoked_at" db:"revoked_at"`
	ExpiresAt time.Time        `json:"expires_at" db:"expires_at"`
}

type AdminTokenRevocation struct {
	TokenRevocation
	UserName *string `json:"user_name" db:"user_name"`
}

func (db *Database) revoke(
	ctx context.Context,
	execer sqlx.ExecerContext,
	kind RevocationKind,
	reason RevocationReason,
	subjects string,
	args ...any,
) error {
	query := fmt.Sprintf(
		`
		WITH subjects AS (%s)
		INSERT INTO token_revocations (subject, kind, user_id, reason, revoked_at, expires_at)
		SELECT DISTINCT ON (subject)
			subject, $%d, user_id, $%d, NOW(),
			NOW() + make_interval(secs => GREATEST(3600, (SELECT COALESCE(MAX(token_duration), 0) FROM applications)))
		FROM subjects
		ON CONFLICT (subject) DO UPDATE SET
			kind = EXCLUDED.kind,
			user_id = EXCLUDED.user_id,
			reason = EXCLUDED.reason,
			revoked_at = EXCLUDED.revoked_at,
			expires_at = EXCLUDED.expires_at
		`,
		subjects,
		len(args)+1,
		len(args)+2,
	)

	_, err := execer.ExecContext(ctx, query, append(args, kind, reason)...)
	return err
}

func (db *Database) RevokeSessionFamily(ctx context.Context, userID string, familyID string, reason RevocationReason) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE family_id = $1`, familyID); err != nil {
		return err
	}

	err = db.revoke(
		ctx,
		tx,
		RevocationKindSession,
		reason,
		`SELECT $1::VARCHAR AS subject, $2::VARCHAR AS user_id`,
		familyID,
		userID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) RevokeUserSessions(ctx context.Context, userID string, reason RevocationReason) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1 OR impersonator_id = $1`, userID); err != nil {
		return err
	}

	if err := db.revokeUser(ctx, tx, userID, reason); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) RevokeUserTokens(ctx context.Context, userID string, reason RevocationReason) error {
	return db.revokeUser(ctx, db.DB(), userID, reason)
}

func (db *Database) revokeUser(ctx context.Context, execer sqlx.ExecerContext, userID string, reason RevocationReason) error {
	return db.revoke(
		ctx,
		execer,
		RevocationKindUser,
		reason,
		`SELECT $1::VARCHAR AS subject, $1::VARCHAR AS user_id`,
		userID,
	)
}

func (db *Database) RevokeRoleTokens(ctx context.Context, roleID string, reason RevocationReason) error {
	return db.revokeRole(ctx, db.DB(), roleID, reason)
}

func (db *Database) revokeRole(ctx context.Context, execer sqlx.ExecerContext, roleID string, reason RevocationReason) error {
	return db.revoke(
		ctx,
		execer,
		RevocationKindUser,
		reason,
		`SELECT user_id AS subject, user_id FROM role_bindings WHERE role_id = $1`,
		roleID,
	)
}

func (db *Database) RevokeAllTokens(ctx context.Context, reason RevocationReason) error {
	return db.revokeAll(ctx, db.DB(), reason)
}

func (db *Database) revokeAll(ctx context.Context, execer sqlx.ExecerContext, reason RevocationReason) error {
	return db.revoke(
		ctx,
		execer,
		RevocationKindGlobal,
		reason,
		`SELECT $1::VARCHAR AS subject, NULL::VARCHAR AS user_id`,
		GlobalRevocationSubject,
	)
}

func (db *Database) GetTokenRevocationsSince(ctx context.Context, since time.Time) ([]TokenRevocation, error) {
	return list[TokenRevocation](
		ctx,
		db.DB(),
		`SELECT * FROM token_revocations WHERE revoked_at > $1 AND expires_at > NOW()`,
		since,
	)
}

func (db *Database) GetTokenRevocations(ctx context.Context, page int, limit int) []AdminTokenRevocation {
	revocations, err := list[AdminTokenRevocation](
		ctx,
		db.DB(),
		`
		SELECT token_revocations.*, users.name AS user_name
		FROM token_revocations
		LEFT JOIN users ON users.id = token_revocations.user_id
		WHERE token_revocations.expires_at > NOW()
		ORDER BY token_revocations.revoked_at DESC
		LIMIT $1 OFFSET $2
		`,
		limit, (page-1)*limit,
	)

	if err != nil {
		slog.Error("failed to get token revocations", "err", err)
		return []AdminTokenRevocation{}
	}

	return revocations
}

func (db *Database) DeleteExpiredTokenRevocations(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(ctx, `DELETE FROM token_revocations WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
