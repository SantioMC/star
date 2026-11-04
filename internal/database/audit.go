package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

const AuditLogKey = "aud"

type AuditLog struct {
	ID          string      `json:"id"`
	Action      string      `json:"action"`
	ActorID     *string     `json:"actor_id" db:"actor_id"`
	ActorName   *string     `json:"actor_name" db:"actor_name"`
	TargetType  *string     `json:"target_type" db:"target_type"`
	TargetID    *string     `json:"target_id" db:"target_id"`
	TargetName  *string     `json:"target_name" db:"target_name"`
	Application *string     `json:"application_id" db:"application_id"`
	Provider    *string     `json:"provider_id" db:"provider_id"`
	Address     *string     `json:"ip_address" db:"ip_address"`
	AddressKey  *string     `json:"-" db:"ip_hash"`
	UserAgent   *string     `json:"user_agent" db:"user_agent"`
	Metadata    models.JSON `json:"metadata"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

type AuditLogFilter struct {
	UserID      *string
	Application *string
	Provider    *string
	Subjects    []string
	AddressKey  *string
	Actions     []string
	Since       *time.Time
	Until       *time.Time
	Before      *Cursor
	Limit       int
}

func (db *Database) HashAuditAddress(address string) string {
	return cryptography.HMAC("audit:"+address, db.config.AppToken)
}

func (db *Database) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	var address *string
	var addressKey *string

	if log.Address != nil && *log.Address != "" {
		encrypted, err := cryptography.EncryptAES(*log.Address, db.config.AppToken)
		if err != nil {
			return err
		}

		address = &encrypted
		addressKey = new(db.HashAuditAddress(*log.Address))
	}

	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO audit_logs (
			id, action, actor_id, actor_name, target_type, target_id, target_name,
			application_id, provider_id, ip_address, ip_hash, user_agent, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
		log.ID,
		log.Action,
		log.ActorID,
		log.ActorName,
		log.TargetType,
		log.TargetID,
		log.TargetName,
		log.Application,
		log.Provider,
		address,
		addressKey,
		log.UserAgent,
		log.Metadata,
		log.CreatedAt,
	)

	return duplicate(err)
}

func (db *Database) GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, error) {
	args := []any{}
	bind := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	conditions := []string{"TRUE"}

	if len(filter.Actions) > 0 {
		conditions = append(conditions, "action = ANY("+bind(pq.StringArray(filter.Actions))+")")
	}

	if filter.Application != nil {
		conditions = append(conditions, "application_id = "+bind(*filter.Application))
	}

	if filter.Provider != nil {
		conditions = append(conditions, "provider_id = "+bind(*filter.Provider))
	}

	if filter.AddressKey != nil {
		conditions = append(conditions, "ip_hash = "+bind(*filter.AddressKey))
	}

	if filter.Since != nil {
		conditions = append(conditions, "created_at >= "+bind(*filter.Since))
	}

	if filter.Until != nil {
		conditions = append(conditions, "created_at <= "+bind(*filter.Until))
	}

	if filter.Before != nil {
		createdAt := bind(filter.Before.CreatedAt)
		id := bind(filter.Before.ID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < (%s, %s)", createdAt, id))
	}

	subjects := filter.Subjects

	if filter.UserID != nil {
		if len(subjects) > 0 {
			matches := bind(pq.StringArray(subjects))
			conditions = append(conditions, fmt.Sprintf(
				"(actor_id = ANY(%[1]s) OR target_id = ANY(%[1]s))",
				matches,
			))
		}

		subjects = []string{*filter.UserID}
	}

	where := strings.Join(conditions, " AND ")
	limit := bind(filter.Limit)

	branch := func(condition string) string {
		return fmt.Sprintf(
			`SELECT * FROM audit_logs WHERE %s AND %s ORDER BY created_at DESC, id DESC LIMIT %s`,
			condition, where, limit,
		)
	}

	query := branch("TRUE")

	if len(subjects) > 0 {
		matches := bind(pq.StringArray(subjects))
		query = fmt.Sprintf(
			`(%s) UNION (%s) ORDER BY created_at DESC, id DESC LIMIT %s`,
			branch("actor_id = ANY("+matches+")"),
			branch("target_id = ANY("+matches+")"),
			limit,
		)
	}

	logs, err := list[AuditLog](ctx, db.DB(), query, args...)
	if err != nil {
		return nil, err
	}

	for idx := range logs {
		if logs[idx].Address == nil {
			continue
		}

		address, err := cryptography.DecryptAES(*logs[idx].Address, db.config.AppToken)
		if err != nil {
			return nil, err
		}

		logs[idx].Address = &address
	}

	return logs, nil
}

func auditPartition(month time.Time) string {
	return month.Format("audit_logs_2006_01")
}

func (db *Database) EnsureAuditPartitions(ctx context.Context, now time.Time) error {
	month := time.Date(now.UTC().Year(), now.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	failures := []error{}

	for offset := range 3 {
		from := month.AddDate(0, offset, 0)
		to := from.AddDate(0, 1, 0)

		_, err := db.DB().ExecContext(ctx, fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs FOR VALUES FROM ('%s') TO ('%s')`,
			auditPartition(from),
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
		))

		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", auditPartition(from), err))
		}
	}

	return errors.Join(failures...)
}

func (db *Database) CountDefaultAuditLogs(ctx context.Context) (int, error) {
	return sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM audit_logs_default`)
}

func (db *Database) DropExpiredAuditPartitions(ctx context.Context, cutoff time.Time) (int, error) {
	partitions, err := list[string](
		ctx,
		db.DB(),
		`
		SELECT child.relname FROM pg_inherits
		JOIN pg_class child ON child.oid = pg_inherits.inhrelid
		WHERE pg_inherits.inhparent = 'audit_logs'::regclass
		`,
	)

	if err != nil {
		return 0, err
	}

	dropped := 0
	for _, partition := range partitions {
		month, err := time.Parse("audit_logs_2006_01", partition)
		if err != nil || month.AddDate(0, 1, 0).After(cutoff) {
			continue
		}

		if _, err := db.DB().ExecContext(ctx, `DROP TABLE IF EXISTS `+auditPartition(month)); err != nil {
			return dropped, err
		}

		dropped++
	}

	return dropped, nil
}

func (db *Database) DeleteExpiredAuditLogs(ctx context.Context, cutoff time.Time, batch int) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		DELETE FROM audit_logs
		WHERE (created_at, id) IN (
			SELECT created_at, id FROM audit_logs
			WHERE created_at < $1
			ORDER BY created_at
			LIMIT $2
		)
		`,
		cutoff,
		batch,
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
