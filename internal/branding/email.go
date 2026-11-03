package branding

import (
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/database"
)

func optional(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func ForEmail(branding *database.Branding, serverURL string) emails.Branding {
	return emails.Branding{
		Name:             branding.Name,
		AccentColor:      branding.AccentColor,
		AccentForeground: Foreground(branding.AccentColor),
		HeadingFont:      branding.HeadingFont,
		BodyFont:         branding.BodyFont,
		FontURL:          FontURL(branding),
		LogoURL:          optional(LogoURL(branding, serverURL)),
		SupportEmail:     optional(branding.SupportEmail),
		PrivacyURL:       optional(branding.PrivacyURL),
		TermsURL:         optional(branding.TermsURL),
	}
}
