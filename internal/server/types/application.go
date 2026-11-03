package types

import "github.com/tubblelive/star/internal/database/models"

type CreateApplicationRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=64"`
	Kind           string  `json:"kind" validate:"required"`
	ApplicationURL string  `json:"url" validate:"required,max=255"`
	ClientID       *string `json:"client_id" validate:"omitempty"`
	Namespace      *string `json:"namespace" validate:"omitempty,max=32"`
}

type UpdateApplicationRequest struct {
	Name             *string         `json:"name" validate:"omitempty,min=1,max=64"`
	Description      *string         `json:"description" validate:"omitempty"`
	Kind             *string         `json:"kind" validate:"omitempty"`
	Consent          *string         `json:"consent" validate:"omitempty"`
	ApplicationURL   *string         `json:"url" validate:"omitempty,max=255"`
	ClientID         *string         `json:"client_id" validate:"omitempty"`
	Namespace        *string         `json:"namespace" validate:"omitempty,max=32"`
	TokenDuration    *models.Seconds `json:"token_duration" validate:"omitempty"`
	SessionDuration  *models.Seconds `json:"session_duration" validate:"omitempty"`
	RedirectURIs     []string        `json:"redirect_uris" validate:"omitempty,max=32"`
	DevelopmentMode  *bool           `json:"development_mode" validate:"omitempty"`
	PrivilegedScopes *bool           `json:"privileged_scopes" validate:"omitempty"`
	SigningAlgorithm *string         `json:"signing_algorithm" validate:"omitempty"`
}

type ApplicationSecretResponse struct {
	ID           string  `json:"id"`
	ClientSecret *string `json:"client_secret,omitempty"`
}

type CreateApplicationScopeRequest struct {
	Name        string `json:"name" validate:"required,max=128"`
	Description string `json:"description" validate:"required,max=255"`
	Requestable bool   `json:"requestable"`
}

type UpdateApplicationScopeRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=128"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Requestable *bool   `json:"requestable" validate:"omitempty"`
}

type CreateApplicationAuthorizationRequest struct {
	ApplicationID       string  `json:"application_id" validate:"required"`
	Scope               string  `json:"scope" validate:"required"`
	CodeChallenge       *string `json:"code_challenge" validate:"omitempty"`
	CodeChallengeMethod *string `json:"code_challenge_method" validate:"omitempty"`
	RedirectURI         string  `json:"redirect_uri" validate:"required"`
	State               *string `json:"state" validate:"omitempty"`
	Nonce               *string `json:"nonce" validate:"omitempty"`
	Consent             bool    `json:"consent"`
}

type CreateApplicationAuthorizationResponse struct {
	Code           string   `json:"code"`
	State          *string  `json:"state"`
	Scopes         []string `json:"scopes"`
	ScopedModified bool     `json:"scoped_modified"`
}
