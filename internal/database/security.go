package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

type SecurityAlertKind string
type SecurityAlertSeverity string

const SecurityAlertKey = "sec"

const (
	SecurityAlertASessionReused        SecurityAlertKind = "authentication.session.already_rotated"
	SecurityAlertAuthorizationCodeUsed SecurityAlertKind = "authorization.code.already_used"
	SecurityAlertDeviceCodeUsed        SecurityAlertKind = "authorization.device_code.already_used"
	SecurityAlertPasskeyCloned         SecurityAlertKind = "authentication.passkey.possibly_cloned"
	SecurityAlertMFABruteForce         SecurityAlertKind = "authentication.mfa.repeated_failures"
	SecurityAlertPasswordBruteForce    SecurityAlertKind = "authentication.password.repeated_failures"

	SecurityAlertSeverityHigh   SecurityAlertSeverity = "high"
	SecurityAlertSeverityMedium SecurityAlertSeverity = "medium"
	SecurityAlertSeverityLow    SecurityAlertSeverity = "low"
)

type SecurityAlert struct {
	ID          string                `json:"id"`
	Kind        SecurityAlertKind     `json:"kind"`
	Severity    SecurityAlertSeverity `json:"severity"`
	Message     string                `json:"message"`
	Data        models.JSON           `json:"data"`
	UserID      *string               `json:"user_id" db:"user_id"`
	Application *string               `json:"application" db:"application_id"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time            `json:"resolved_at" db:"resolved_at"`
	ResolvedBy  *string               `json:"resolved_by" db:"resolved_by"`
}

func (db *Database) GetSecurityAlertByID(ctx context.Context, id string) *SecurityAlert {
	securityAlert, err := sqlx.OneContext[SecurityAlert](
		ctx,
		db.DB(),
		`SELECT * FROM security_alerts WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get security alert by id", "err", err)
		return nil
	}

	return &securityAlert
}

func (db *Database) GetSecurityAlerts(ctx context.Context, unresolved bool, page int, perPage int) []SecurityAlert {
	securityAlerts, err := list[SecurityAlert](
		ctx,
		db.DB(),
		`
		SELECT * FROM security_alerts
		WHERE $1 = FALSE OR resolved_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
		`,
		unresolved,
		perPage,
		(page-1)*perPage,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []SecurityAlert{}
	} else if err != nil {
		slog.Error("failed to get security alerts", "err", err)
		return []SecurityAlert{}
	}

	return securityAlerts
}

func (db *Database) ResolveSecurityAlert(ctx context.Context, id string, userID string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE security_alerts
		SET resolved_at = $2, resolved_by = $3
		WHERE id = $1 AND resolved_at IS NULL
		`,
		id,
		time.Now(),
		userID,
	)

	if err != nil {
		return false, err
	}

	resolved, err := result.RowsAffected()
	return resolved > 0, err
}

func (db *Database) GetSecurityAlertsByUserID(ctx context.Context, userID string) []SecurityAlert {
	securityAlerts, err := list[SecurityAlert](
		ctx,
		db.DB(),
		`SELECT * FROM security_alerts WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if err != nil {
		slog.Error("failed to get user security alerts", "err", err)
		return []SecurityAlert{}
	}

	return securityAlerts
}

func (db *Database) CreateSecurityAlert(ctx context.Context, securityAlert *SecurityAlert) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO security_alerts (
			id, kind, severity, message, data, user_id, application_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		securityAlert.ID,
		securityAlert.Kind,
		securityAlert.Severity,
		securityAlert.Message,
		securityAlert.Data,
		securityAlert.UserID,
		securityAlert.Application,
		securityAlert.CreatedAt,
	)

	return err
}

func (db *Database) CountUnresolvedSecurityAlerts(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM security_alerts WHERE resolved_at IS NULL`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
