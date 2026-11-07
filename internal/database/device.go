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

const DeviceAuthorizationKey = "dev"

type DeviceAuthorizationStatus string

const (
	DeviceAuthorizationPending  DeviceAuthorizationStatus = "pending"
	DeviceAuthorizationApproved DeviceAuthorizationStatus = "approved"
	DeviceAuthorizationDenied   DeviceAuthorizationStatus = "denied"
)

type DeviceAuthorization struct {
	ID            string  `json:"id"`
	ApplicationID string  `json:"application_id" db:"application_id"`
	UserID        *string `json:"user_id" db:"user_id"`

	DeviceCode string `json:"device_code" db:"device_code"`
	UserCode   string `json:"user_code" db:"user_code"`

	Scopes pq.StringArray            `json:"scopes"`
	Status DeviceAuthorizationStatus `json:"status"`

	Device      string  `json:"device"`
	Geolocation *string `json:"geolocation"`

	PollInterval int     `json:"poll_interval" db:"poll_interval"`
	FamilyID     *string `json:"family_id" db:"family_id"`

	AuthenticatedAt *time.Time `json:"authenticated_at" db:"authenticated_at"`
	LastPolledAt    time.Time  `json:"last_polled_at" db:"last_polled_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt       time.Time  `json:"expires_at" db:"expires_at"`
	RedeemedAt      *time.Time `json:"redeemed_at" db:"redeemed_at"`
}

func (db *Database) GetDeviceAuthorizationByID(ctx context.Context, id string) *DeviceAuthorization {
	authorization, err := sqlx.OneContext[DeviceAuthorization](
		ctx,
		db.DB(),
		`SELECT * FROM device_authorizations WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get device authorization by id", "err", err)
		return nil
	}

	return &authorization
}

func (db *Database) GetDeviceAuthorizationByDeviceCode(ctx context.Context, deviceCode string) *DeviceAuthorization {
	hashed := cryptography.HMAC(deviceCode, db.config.AppToken)

	authorization, err := sqlx.OneContext[DeviceAuthorization](
		ctx,
		db.DB(),
		`SELECT * FROM device_authorizations WHERE device_code = $1 LIMIT 1`,
		hashed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get device authorization by device code", "err", err)
		return nil
	}

	return &authorization
}

func (db *Database) GetPendingDeviceAuthorization(ctx context.Context, userCode string) *DeviceAuthorization {
	hashed := cryptography.HMAC(userCode, db.config.AppToken)

	authorization, err := sqlx.OneContext[DeviceAuthorization](
		ctx,
		db.DB(),
		`SELECT * FROM device_authorizations WHERE user_code = $1 AND status = $2 AND expires_at > $3 LIMIT 1`,
		hashed,
		DeviceAuthorizationPending,
		time.Now(),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get pending device authorization", "err", err)
		return nil
	}

	return &authorization
}

func (db *Database) GetDeviceAuthorizationsByUserID(ctx context.Context, userID string) []DeviceAuthorization {
	authorizations, err := list[DeviceAuthorization](
		ctx,
		db.DB(),
		`SELECT * FROM device_authorizations WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if err != nil {
		slog.Error("failed to get user device authorizations", "err", err)
		return []DeviceAuthorization{}
	}

	return authorizations
}

func (db *Database) CreateDeviceAuthorization(ctx context.Context, authorization *DeviceAuthorization) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO device_authorizations (
			id, application_id, device_code, user_code, scopes, status,
			device, geolocation, poll_interval, last_polled_at, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
		authorization.ID,
		authorization.ApplicationID,
		cryptography.HMAC(authorization.DeviceCode, db.config.AppToken),
		cryptography.HMAC(authorization.UserCode, db.config.AppToken),
		authorization.Scopes,
		authorization.Status,
		authorization.Device,
		authorization.Geolocation,
		authorization.PollInterval,
		authorization.LastPolledAt,
		authorization.CreatedAt,
		authorization.ExpiresAt,
	)

	return duplicate(err)
}

func (db *Database) ApproveDeviceAuthorization(ctx context.Context, id string, userID string, scopes []string, authenticatedAt time.Time) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE device_authorizations
		SET status = $1, user_id = $2, scopes = $3, authenticated_at = $4
		WHERE id = $5 AND status = $6 AND expires_at > $7
		`,
		DeviceAuthorizationApproved,
		userID,
		pq.StringArray(scopes),
		authenticatedAt,
		id,
		DeviceAuthorizationPending,
		time.Now(),
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (db *Database) DenyDeviceAuthorization(ctx context.Context, id string, userID string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE device_authorizations
		SET status = $1, user_id = $2
		WHERE id = $3 AND status = $4 AND expires_at > $5
		`,
		DeviceAuthorizationDenied,
		userID,
		id,
		DeviceAuthorizationPending,
		time.Now(),
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (db *Database) PollDeviceAuthorization(ctx context.Context, id string) (bool, error) {
	now := time.Now()

	result, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE device_authorizations
		SET last_polled_at = $1
		WHERE id = $2 AND last_polled_at + make_interval(secs => poll_interval - 1) <= $1
		`,
		now,
		id,
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if affected > 0 {
		return true, nil
	}

	_, err = db.DB().ExecContext(
		ctx,
		`UPDATE device_authorizations SET last_polled_at = $1, poll_interval = poll_interval + 5 WHERE id = $2`,
		now,
		id,
	)

	return false, err
}

func (db *Database) RedeemDeviceAuthorization(ctx context.Context, id string, familyID string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`UPDATE device_authorizations SET redeemed_at = $1, family_id = $2 WHERE id = $3 AND status = $4 AND redeemed_at IS NULL`,
		time.Now(),
		familyID,
		id,
		DeviceAuthorizationApproved,
	)

	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (db *Database) DeleteDeadDeviceAuthorizations(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM device_authorizations WHERE expires_at < $1 OR redeemed_at < $1`,
		time.Now().Add(-time.Hour),
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
