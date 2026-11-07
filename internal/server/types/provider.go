package types

import "time"

type CreateProviderRequest struct {
	Name             string   `json:"name" validate:"required,min=1,max=64"`
	Icon             *string  `json:"icon" validate:"omitempty,max=64"`
	Order            *int     `json:"order" validate:"omitempty,min=0"`
	ClientID         string   `json:"client_id" validate:"required,max=255"`
	ClientSecret     string   `json:"client_secret" validate:"required,max=255"`
	Scopes           []string `json:"scopes" validate:"omitempty"`
	OIDCUrl          *string  `json:"oidc_url" validate:"omitempty,https_url"`
	AuthorizationURL *string  `json:"authorization_url" validate:"omitempty,https_url"`
	TokenURL         *string  `json:"token_url" validate:"omitempty,https_url"`
	UserinfoURL      *string  `json:"userinfo_url" validate:"omitempty,https_url"`
	TrustEmail       bool     `json:"trust_email"`
	Enabled          *bool    `json:"enabled"`
}

type UpdateProviderRequest struct {
	Name             *string   `json:"name" validate:"omitempty,min=1,max=64"`
	Icon             *string   `json:"icon" validate:"omitempty,max=64"`
	Order            *int      `json:"order" validate:"omitempty,min=0"`
	ClientID         *string   `json:"client_id" validate:"omitempty,max=255"`
	ClientSecret     *string   `json:"client_secret" validate:"omitempty,min=1,max=255"`
	Scopes           *[]string `json:"scopes" validate:"omitempty"`
	OIDCUrl          *string   `json:"oidc_url" validate:"omitempty,https_url"`
	AuthorizationURL *string   `json:"authorization_url" validate:"omitempty,https_url"`
	TokenURL         *string   `json:"token_url" validate:"omitempty,https_url"`
	UserinfoURL      *string   `json:"userinfo_url" validate:"omitempty,https_url"`
	TrustEmail       *bool     `json:"trust_email" validate:"omitempty"`
	Enabled          *bool     `json:"enabled" validate:"omitempty"`
}

type ProviderLoginState struct {
	Provider  string    `json:"provider"`
	UserID    string    `json:"user_id,omitempty"`
	FamilyID  string    `json:"family_id,omitempty"`
	State     string    `json:"state"`
	Verifier  string    `json:"verifier"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
}

type CompleteProviderLoginRequest struct {
	Flow string `json:"flow" validate:"required,max=64"`
}
