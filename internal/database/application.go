package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

type ApplicationKind string

const ApplicationKey = "app"

const (
	ApplicationKindPublic       ApplicationKind = "public"
	ApplicationKindConfidential ApplicationKind = "confidential"
)

var ApplicationKinds = []ApplicationKind{
	ApplicationKindPublic, ApplicationKindConfidential,
}

type ApplicationConsent string

const (
	ApplicationConsentDefault  ApplicationConsent = "default"
	ApplicationConsentExplicit ApplicationConsent = "explicit"
	ApplicationConsentImplicit ApplicationConsent = "implicit"
)

var ApplicationConsents = []ApplicationConsent{
	ApplicationConsentDefault, ApplicationConsentExplicit, ApplicationConsentImplicit,
}

type Application struct {
	ID               string             `json:"id"`
	Kind             ApplicationKind    `json:"kind"`
	Name             string             `json:"name"`
	Namespace        string             `json:"namespace"`
	ApplicationURL   string             `json:"application_url" db:"application_url"`
	ClientID         string             `json:"client_id" db:"client_id"`
	ClientSecret     *string            `json:"-" db:"client_secret"`
	Internal         bool               `json:"internal"`
	Consent          ApplicationConsent `json:"consent"`
	TokenDuration    models.Seconds     `json:"token_duration" db:"token_duration"`
	SessionDuration  models.Seconds     `json:"session_duration" db:"session_duration"`
	RedirectURIs     pq.StringArray     `json:"redirect_uris" db:"redirect_uris"`
	DevelopmentMode  bool               `json:"development_mode" db:"development_mode"`
	PrivilegedScopes bool               `json:"privileged_scopes" db:"privileged_scopes"`
	SigningAlgorithm KeyAlgorithm       `json:"signing_algorithm" db:"signing_algorithm"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
}

func CreateApplicationSecret() string {
	return cryptography.CreateSecureString(48)
}

func (db *Database) HashApplicationSecret(secret string) string {
	return cryptography.HMAC(secret, db.config.AppToken)
}

func (db *Database) GetApplicationByID(ctx context.Context, id string) *Application {
	application, err := sqlx.OneContext[Application](
		ctx,
		db.DB(),
		`SELECT * FROM applications WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application by id", "err", err)
		return nil
	}

	return &application
}

func (db *Database) GetApplicationByClientID(ctx context.Context, clientID string) *Application {
	application, err := sqlx.OneContext[Application](
		ctx,
		db.DB(),
		`SELECT * FROM applications WHERE client_id = $1 LIMIT 1`,
		clientID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application by client id", "err", err)
		return nil
	}

	return &application
}

func (db *Database) GetApplicationByName(ctx context.Context, name string) *Application {
	application, err := sqlx.OneContext[Application](
		ctx,
		db.DB(),
		`SELECT * FROM applications WHERE name = $1 LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application by name", "err", err)
		return nil
	}

	return &application
}

func (db *Database) GetApplicationByNamespace(ctx context.Context, namespace string) *Application {
	application, err := sqlx.OneContext[Application](
		ctx,
		db.DB(),
		`SELECT * FROM applications WHERE namespace = $1 LIMIT 1`,
		namespace,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application by namespace", "err", err)
		return nil
	}

	return &application
}

func (db *Database) CreateApplication(ctx context.Context, application *Application) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO applications (
			id, kind, name, namespace, application_url, client_id,
			client_secret, internal, token_duration,
			session_duration, redirect_uris, development_mode,
			created_at, consent, privileged_scopes, signing_algorithm
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		`,
		application.ID,
		application.Kind,
		application.Name,
		application.Namespace,
		application.ApplicationURL,
		application.ClientID,
		application.ClientSecret,
		application.Internal,
		application.TokenDuration,
		application.SessionDuration,
		application.RedirectURIs,
		application.DevelopmentMode,
		application.CreatedAt,
		application.Consent,
		application.PrivilegedScopes,
		application.SigningAlgorithm,
	)

	return err
}

func (db *Database) UpdateApplication(ctx context.Context, application *Application) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE applications
		SET name = $2, application_url = $3, client_id = $4,
		client_secret = $5, internal = $6, token_duration = $7,
		session_duration = $8, redirect_uris = $9, development_mode = $10,
		kind = $11, namespace = $12, consent = $13, privileged_scopes = $14,
		signing_algorithm = $15
		WHERE id = $1
		`,
		application.ID,
		application.Name,
		application.ApplicationURL,
		application.ClientID,
		application.ClientSecret,
		application.Internal,
		application.TokenDuration,
		application.SessionDuration,
		application.RedirectURIs,
		application.DevelopmentMode,
		application.Kind,
		application.Namespace,
		application.Consent,
		application.PrivilegedScopes,
		application.SigningAlgorithm,
	)

	return err
}

func (db *Database) DeleteApplicationByID(ctx context.Context, id string) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = db.revoke(
		ctx,
		tx,
		RevocationKindSession,
		RevocationReasonAppDeleted,
		`DELETE FROM sessions WHERE application_id = $1 RETURNING family_id AS subject, user_id`,
		id,
	)

	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM applications WHERE id = $1`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) GetApplications(ctx context.Context) []Application {
	applications, err := list[Application](
		ctx,
		db.DB(),
		`SELECT * FROM applications ORDER BY created_at DESC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Application{}
	} else if err != nil {
		slog.Error("failed to get applications", "err", err)
		return []Application{}
	}

	return applications
}

func (db *Database) GetLongestTokenDuration(ctx context.Context) (time.Duration, error) {
	seconds, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`SELECT COALESCE(MAX(token_duration), 0) FROM applications`,
	)

	if err != nil {
		return 0, err
	}

	return time.Duration(seconds) * time.Second, nil
}

func (db *Database) GetApplicationCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM applications`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
