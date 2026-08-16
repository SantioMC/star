package types

type CreateProviderRequest struct {
	Name             string   `json:"name" validate:"required,min=1,max=64"`
	Icon             *string  `json:"icon" validate:"omitempty,max=64"`
	Order            *int     `json:"order" validate:"omitempty,min=0"`
	ClientID         string   `json:"client_id" validate:"required,max=255"`
	ClientSecret     string   `json:"client_secret" validate:"required,max=255"`
	Scopes           []string `json:"scopes" validate:"omitempty"`
	OIDCUrl          *string  `json:"oidc_url" validate:"omitempty,http_url"`
	AuthorizationURL *string  `json:"authorization_url" validate:"omitempty,http_url"`
	TokenURL         *string  `json:"token_url" validate:"omitempty,http_url"`
	UserinfoURL      *string  `json:"userinfo_url" validate:"omitempty,http_url"`
}

type UpdateProviderRequest struct {
	Name             *string   `json:"name" validate:"omitempty,min=1,max=64"`
	Icon             *string   `json:"icon" validate:"omitempty,max=64"`
	Order            *int      `json:"order" validate:"omitempty,min=0"`
	ClientID         *string   `json:"client_id" validate:"omitempty,max=255"`
	ClientSecret     *string   `json:"client_secret" validate:"omitempty,max=255"`
	Scopes           *[]string `json:"scopes" validate:"omitempty"`
	OIDCUrl          *string   `json:"oidc_url" validate:"omitempty,http_url"`
	AuthorizationURL *string   `json:"authorization_url" validate:"omitempty,http_url"`
	TokenURL         *string   `json:"token_url" validate:"omitempty,http_url"`
	UserinfoURL      *string   `json:"userinfo_url" validate:"omitempty,http_url"`
}
