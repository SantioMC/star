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

const ApplicationScopeKey = "scope"

type ApplicationScope struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id" db:"application_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Requestable   bool      `json:"requestable"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type OwnedApplicationScope struct {
	ApplicationScope
	ApplicationName      string `json:"application_name" db:"application_name"`
	ApplicationNamespace string `json:"application_namespace" db:"application_namespace"`
}

func (db *Database) GetApplicationScopes(ctx context.Context, applicationID string) []ApplicationScope {
	scopes, err := sqlx.ListContext[ApplicationScope](
		ctx,
		db.DB(),
		`SELECT * FROM application_scopes WHERE application_id = $1 ORDER BY name ASC`,
		applicationID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []ApplicationScope{}
	} else if err != nil {
		slog.Error("failed to get application scopes", "err", err)
		return []ApplicationScope{}
	}

	return scopes
}

func (db *Database) GetAllApplicationScopes(ctx context.Context) []ApplicationScope {
	scopes, err := sqlx.ListContext[ApplicationScope](
		ctx,
		db.DB(),
		`SELECT * FROM application_scopes ORDER BY name ASC`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []ApplicationScope{}
	} else if err != nil {
		slog.Error("failed to get all application scopes", "err", err)
		return []ApplicationScope{}
	}

	return scopes
}

func (db *Database) GetApplicationScopeByID(ctx context.Context, id string) *ApplicationScope {
	scope, err := sqlx.OneContext[ApplicationScope](
		ctx,
		db.DB(),
		`SELECT * FROM application_scopes WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application scope by id", "err", err)
		return nil
	}

	return &scope
}

func (db *Database) GetApplicationScopeByName(ctx context.Context, name string) *ApplicationScope {
	scope, err := sqlx.OneContext[ApplicationScope](
		ctx,
		db.DB(),
		`SELECT * FROM application_scopes WHERE name = $1 LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get application scope by name", "err", err)
		return nil
	}

	return &scope
}

func (db *Database) GetApplicationScopesByNames(ctx context.Context, names []string) []OwnedApplicationScope {
	if len(names) == 0 {
		return []OwnedApplicationScope{}
	}

	scopes, err := sqlx.ListContext[OwnedApplicationScope](
		ctx,
		db.DB(),
		`
		SELECT scope.*, app.name AS application_name, app.namespace AS application_namespace
		FROM application_scopes scope
		JOIN applications app ON app.id = scope.application_id
		WHERE scope.name = ANY($1)
		`,
		pq.StringArray(names),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []OwnedApplicationScope{}
	} else if err != nil {
		slog.Error("failed to get application scopes by names", "err", err)
		return []OwnedApplicationScope{}
	}

	return scopes
}

func (db *Database) CreateApplicationScope(ctx context.Context, scope *ApplicationScope) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO application_scopes (
			id, application_id, name, description, requestable, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		`,
		scope.ID,
		scope.ApplicationID,
		scope.Name,
		scope.Description,
		scope.Requestable,
		scope.CreatedAt,
	)

	return err
}

func (db *Database) UpdateApplicationScope(ctx context.Context, scope *ApplicationScope) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE application_scopes SET name = $2, description = $3, requestable = $4 WHERE id = $1`,
		scope.ID,
		scope.Name,
		scope.Description,
		scope.Requestable,
	)

	return err
}

func (db *Database) DeleteApplicationScopeByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM application_scopes WHERE id = $1`,
		id,
	)

	return err
}
