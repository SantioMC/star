package transformer

import (
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/database"
)

type Branding struct {
	Name         string                  `json:"name"`
	AccentColor  string                  `json:"accent_color"`
	HeadingFont  string                  `json:"heading_font"`
	BodyFont     string                  `json:"body_font"`
	Layout       database.BrandingLayout `json:"layout"`
	LogoURL      *string                 `json:"logo_url"`
	SupportEmail *string                 `json:"support_email"`
	PrivacyURL   *string                 `json:"privacy_url"`
	TermsURL     *string                 `json:"terms_url"`
}

func TransformBranding(value *database.Branding, serverURL string) Branding {
	return Branding{
		Name:         value.Name,
		AccentColor:  value.AccentColor,
		HeadingFont:  value.HeadingFont,
		BodyFont:     value.BodyFont,
		Layout:       value.Layout,
		LogoURL:      branding.LogoURL(value, serverURL),
		SupportEmail: value.SupportEmail,
		PrivacyURL:   value.PrivacyURL,
		TermsURL:     value.TermsURL,
	}
}
