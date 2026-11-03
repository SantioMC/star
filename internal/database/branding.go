package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vinovest/sqlx"
)

type BrandingLayout string

const (
	BrandingLayoutCentered BrandingLayout = "centered"
	BrandingLayoutSplit    BrandingLayout = "split"
	BrandingLayoutMinimal  BrandingLayout = "minimal"
)

type Branding struct {
	ID            int            `db:"id"`
	Name          string         `db:"name"`
	AccentColor   string         `db:"accent_color"`
	HeadingFont   string         `db:"heading_font"`
	BodyFont      string         `db:"body_font"`
	Layout        BrandingLayout `db:"layout"`
	SupportEmail  *string        `db:"support_email"`
	PrivacyURL    *string        `db:"privacy_url"`
	TermsURL      *string        `db:"terms_url"`
	LogoUpdatedAt *time.Time     `db:"logo_updated_at"`
}

type BrandingLogo struct {
	Data      []byte    `db:"logo"`
	Type      string    `db:"logo_type"`
	UpdatedAt time.Time `db:"logo_updated_at"`
}

func (db *Database) GetBranding(ctx context.Context) (*Branding, error) {
	branding, err := sqlx.OneContext[Branding](
		ctx,
		db.DB(),
		`
		SELECT id, name, accent_color, heading_font, body_font, layout, support_email, privacy_url, terms_url, logo_updated_at
		FROM branding
		LIMIT 1
		`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &branding, nil
}

func (db *Database) GetBrandingLogo(ctx context.Context) (*BrandingLogo, error) {
	logo, err := sqlx.OneContext[BrandingLogo](
		ctx,
		db.DB(),
		`SELECT logo, logo_type, logo_updated_at FROM branding WHERE logo IS NOT NULL LIMIT 1`,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &logo, nil
}

func (db *Database) UpdateBranding(ctx context.Context, branding *Branding) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE branding
		SET name = $2, accent_color = $3, heading_font = $4, body_font = $5, layout = $6, support_email = $7, privacy_url = $8, terms_url = $9
		WHERE id = $1
		`,
		branding.ID,
		branding.Name,
		branding.AccentColor,
		branding.HeadingFont,
		branding.BodyFont,
		branding.Layout,
		branding.SupportEmail,
		branding.PrivacyURL,
		branding.TermsURL,
	)

	return err
}

func (db *Database) UpdateBrandingLogo(ctx context.Context, data []byte, contentType string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE branding SET logo = $1, logo_type = $2, logo_updated_at = $3`,
		data,
		contentType,
		time.Now(),
	)

	return err
}

func (db *Database) DeleteBrandingLogo(ctx context.Context) error {
	_, err := db.DB().ExecContext(
		ctx,
		`UPDATE branding SET logo = NULL, logo_type = NULL, logo_updated_at = NULL`,
	)

	return err
}

func (db *Database) BrandingExists(ctx context.Context) bool {
	exists, err := sqlx.OneContext[bool](
		ctx,
		db.DB(),
		`SELECT EXISTS(SELECT 1 FROM branding LIMIT 1)`,
	)

	return err == nil && exists
}

func (db *Database) SaveDefaultBranding(ctx context.Context) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO branding (id, name, accent_color, heading_font, body_font, layout)
		VALUES (1, $1, $2, $3, $4, $5)
		`,
		"Star",
		"#e0a458",
		"Exo",
		"Instrument Sans",
		BrandingLayoutCentered,
	)

	return err
}
