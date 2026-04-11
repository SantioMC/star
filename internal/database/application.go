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

type ApplicationKind string

const (
	ApplicationKindPublic       ApplicationKind = "public"
	ApplicationKindPKCE         ApplicationKind = "pkce"
	ApplicationKindConfidential ApplicationKind = "confidential"
)

var ApplicationKinds = []ApplicationKind{
	ApplicationKindPublic, ApplicationKindPKCE, ApplicationKindConfidential,
}

type Application struct {
	ID              string          `json:"id"`
	Kind            ApplicationKind `json:"kind"`
	Name            string          `json:"name"`
	ApplicationURL  string          `json:"application_url" db:"application_url"`
	ClientID        string          `json:"client_id" db:"client_id"`
	ClientSecret    *string         `json:"client_secret" db:"client_secret"`
	Internal        bool            `json:"internal"`
	TokenDuration   time.Duration   `json:"token_duration" db:"token_duration"`
	SessionDuration time.Duration   `json:"session_duration" db:"session_duration"`
	RedirectURIs    pq.StringArray  `json:"redirect_uris" db:"redirect_uris"`
	DevelopmentMode bool            `json:"development_mode" db:"development_mode"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
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

func (db *Database) CreateApplication(ctx context.Context, application *Application) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO applications (
			id, kind, name, application_url, client_id, 
			client_secret, internal, token_duration, 
			session_duration, redirect_uris, development_mode, 
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
		application.ID,
		application.Kind,
		application.Name,
		application.ApplicationURL,
		application.ClientID,
		application.ClientSecret,
		application.Internal,
		application.TokenDuration,
		application.SessionDuration,
		application.RedirectURIs,
		application.DevelopmentMode,
		application.CreatedAt,
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
		session_duration = $8, redirect_uris = $9, development_mode = $10
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
	)

	return err
}

func (db *Database) DeleteApplicationByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM applications WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) GetApplications(ctx context.Context) []Application {
	applications, err := sqlx.ListContext[Application](
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

func (db *Database) GetApplicationCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM applications`)

	if err != nil {
		return 0, err
	}

	return count, nil
}
