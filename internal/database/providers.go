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

const ProviderKey = "prov"

type Provider struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Icon             string         `json:"icon"`
	Order            int            `json:"order" db:"display_order"`
	ClientID         string         `json:"client_id" db:"client_id"`
	ClientSecret     string         `json:"-" db:"client_secret"`
	Scopes           pq.StringArray `json:"scopes"`
	OIDCUrl          *string        `json:"oidc_url" db:"oidc_url"`
	AuthorizationURL *string        `json:"authorization_url" db:"authorization_url"`
	TokenURL         *string        `json:"token_url" db:"token_url"`
	UserinfoURL      *string        `json:"userinfo_url" db:"userinfo_url"`
	TrustEmail       bool           `json:"trust_email" db:"trust_email"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

func (db *Database) GetProviders(ctx context.Context) []Provider {
	providers, err := list[Provider](
		ctx,
		db.DB(),
		`SELECT * FROM providers ORDER BY display_order, created_at`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []Provider{}
	} else if err != nil {
		slog.Error("failed to get providers", "err", err)
		return []Provider{}
	}

	for idx, provider := range providers {
		if providers[idx].ClientSecret, err = cryptography.DecryptAES(provider.ClientSecret, db.config.AppToken); err != nil {
			slog.Error("failed to decrypt provider client secret", "err", err)
			return []Provider{}
		}
	}

	return providers
}

func (db *Database) GetProviderByID(ctx context.Context, id string) *Provider {
	provider, err := sqlx.OneContext[Provider](
		ctx,
		db.DB(),
		`SELECT * FROM providers WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get provider by id", "err", err)
		return nil
	}

	if provider.ClientSecret, err = cryptography.DecryptAES(provider.ClientSecret, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt provider client secret", "err", err)
		return nil
	}

	return &provider
}

func (db *Database) GetProviderByName(ctx context.Context, name string) *Provider {
	provider, err := sqlx.OneContext[Provider](
		ctx,
		db.DB(),
		`SELECT * FROM providers WHERE name = $1 LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get provider by name", "err", err)
		return nil
	}

	if provider.ClientSecret, err = cryptography.DecryptAES(provider.ClientSecret, db.config.AppToken); err != nil {
		slog.Error("failed to decrypt provider client secret", "err", err)
		return nil
	}

	return &provider
}

func (db *Database) CreateProvider(ctx context.Context, provider *Provider) error {
	clientSecret, err := cryptography.EncryptAES(provider.ClientSecret, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = db.DB().ExecContext(
		ctx,
		`
		INSERT INTO providers (
			id, name, icon, display_order, client_id, client_secret,
			scopes, oidc_url, authorization_url, token_url, userinfo_url,
			trust_email, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`,
		provider.ID,
		provider.Name,
		provider.Icon,
		provider.Order,
		provider.ClientID,
		clientSecret,
		provider.Scopes,
		provider.OIDCUrl,
		provider.AuthorizationURL,
		provider.TokenURL,
		provider.UserinfoURL,
		provider.TrustEmail,
		provider.CreatedAt,
	)

	return err
}

func (db *Database) UpdateProvider(ctx context.Context, provider *Provider) error {
	clientSecret, err := cryptography.EncryptAES(provider.ClientSecret, db.config.AppToken)
	if err != nil {
		return err
	}

	_, err = db.DB().ExecContext(
		ctx,
		`
		UPDATE providers
		SET name = $2, icon = $3, display_order = $4, client_id = $5,
		client_secret = $6, scopes = $7, oidc_url = $8,
		authorization_url = $9, token_url = $10, userinfo_url = $11,
		trust_email = $12
		WHERE id = $1
		`,
		provider.ID,
		provider.Name,
		provider.Icon,
		provider.Order,
		provider.ClientID,
		clientSecret,
		provider.Scopes,
		provider.OIDCUrl,
		provider.AuthorizationURL,
		provider.TokenURL,
		provider.UserinfoURL,
		provider.TrustEmail,
	)

	return err
}

func (db *Database) DeleteProviderByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM providers WHERE id = $1`,
		id,
	)

	return err
}
