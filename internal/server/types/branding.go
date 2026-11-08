package types

import "github.com/tubblelive/star/internal/database"

type UpdateBrandingRequest struct {
	Name         *string                  `json:"name" validate:"omitnil,min=1,max=64"`
	AccentColor  *string                  `json:"accent_color" validate:"omitnil,len=7,hexcolor"`
	HeadingFont  *string                  `json:"heading_font" validate:"omitnil,min=1,max=64"`
	BodyFont     *string                  `json:"body_font" validate:"omitnil,min=1,max=64"`
	Layout       *database.BrandingLayout `json:"layout" validate:"omitnil,oneof=centered split minimal"`
	SupportEmail *string                  `json:"support_email" validate:"omitnil,max=255,email|len=0"`
	PrivacyURL   *string                  `json:"privacy_url" validate:"omitnil,max=2048,https_url|len=0"`
	TermsURL     *string                  `json:"terms_url" validate:"omitnil,max=2048,https_url|len=0"`
}

type EmailPreviewRequest struct {
	Template   string                `json:"template" validate:"required,oneof=verification password-reset password-changed invite test"`
	RemoveLogo bool                  `json:"remove_logo"`
	Branding   UpdateBrandingRequest `json:"branding"`
}

type EmailPreviewTicket struct {
	EmailPreviewRequest
	ExpiresAt int64 `json:"expires_at"`
}

type EmailPreviewResponse struct {
	Ticket string `json:"ticket"`
}
