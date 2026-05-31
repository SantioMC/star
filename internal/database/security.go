package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

type SecurityAlertKind string
type SecurityAlertSeverity string

const SecurityAlertKey = "sec"

const (
	SecurityAlertAuthorizationCodeUsed SecurityAlertKind = "authorization.code.already_used"

	SecurityAlertSeverityHigh   SecurityAlertSeverity = "high"
	SecurityAlertSeverityMedium SecurityAlertSeverity = "medium"
	SecurityAlertSeverityLow    SecurityAlertSeverity = "low"
)

type SecurityAlert struct {
	ID          string                `json:"id"`
	Kind        SecurityAlertKind     `json:"kind"`
	Severity    SecurityAlertSeverity `json:"severity"`
	Message     string                `json:"message"`
	Data        map[string]any        `json:"data"`
	UserID      *string               `json:"user_id"`
	Application *string               `json:"application"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
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

func (db *Database) GetSecurityAlerts(ctx context.Context, page int, perPage int) []SecurityAlert {
	securityAlerts, err := sqlx.ListContext[SecurityAlert](
		ctx,
		db.DB(),
		`SELECT * FROM security_alerts ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		perPage,
		(page-1)*perPage,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get security alerts", "err", err)
		return nil
	}

	return securityAlerts
}

func (db *Database) CreateSecurityAlert(ctx context.Context, securityAlert *SecurityAlert) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO security_alerts (
			id, kind, severity, message, data, user_id, application, created_at
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

func (db *Database) CountSecurityAlerts(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM security_alerts`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
