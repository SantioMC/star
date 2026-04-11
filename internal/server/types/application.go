package types

import "time"

type CreateApplicationRequest struct {
	Name           string  `json:"name" validate:"required"`
	Kind           string  `json:"kind" validate:"required"`
	ApplicationURL string  `json:"url" validate:"required"`
	ClientID       *string `json:"client_id" validate:"omitempty"`
}

type UpdateApplicationRequest struct {
	Name            *string        `json:"name" validate:"omitempty"`
	Description     *string        `json:"description" validate:"omitempty"`
	Kind            *string        `json:"kind" validate:"omitempty"`
	ApplicationURL  *string        `json:"url" validate:"omitempty"`
	ClientID        *string        `json:"client_id" validate:"omitempty"`
	TokenDuration   *time.Duration `json:"token_duration" validate:"omitempty"`
	SessionDuration *time.Duration `json:"session_duration" validate:"omitempty"`
	RedirectURIs    []string       `json:"redirect_uris" validate:"omitempty"`
	DevelopmentMode *bool          `json:"development_mode" validate:"omitempty"`
}

type CreateApplicationAuthorizationRequest struct {
	ApplicationID       string  `json:"application_id" validate:"required"`
	Scope               string  `json:"scope" validate:"required"`
	CodeChallenge       *string `json:"code_challenge" validate:"omitempty"`
	CodeChallengeMethod *string `json:"code_challenge_method" validate:"omitempty"`
	RedirectURI         string  `json:"redirect_uri" validate:"required"`
	State               *string `json:"state" validate:"omitempty"`
	Nonce               *string `json:"nonce" validate:"omitempty"`
}

type CreateApplicationAuthorizationResponse struct {
	Code           string   `json:"code"`
	State          *string  `json:"state"`
	Scopes         []string `json:"scopes"`
	ScopedModified bool     `json:"scoped_modified"`
}
