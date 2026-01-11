package model

import (
	"github.com/tubblelive/star/star/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/datatypes"
)

type AppType string

const (
	AppTypePublic       AppType = "public"
	AppTypeConfidential AppType = "confidential"
)

var KnownAppTypes = []AppType{
	AppTypePublic, AppTypeConfidential,
}

type OAuthProvider struct {
	BaseModel

	Application   *Application `json:"-"`
	ApplicationID string       `json:"application_id"`

	RotateRefreshToken bool    `json:"rotate_refresh_token"`
	ApplicationType    AppType `json:"application_type"`

	ClientID string `json:"client_id"`
	Secret   string `json:"client_secret" encrypt:"true"`

	RedirectURIs     datatypes.JSONType[[]string] `json:"redirect_uris"`
	RedirectURIGlobs datatypes.JSONType[[]string] `json:"redirect_uri_globs"`

	ResponseTypes   datatypes.JSONType[[]oidc.ResponseType] `json:"response_types"`
	GrantTypes      datatypes.JSONType[[]oidc.GrantType]    `json:"grant_types"`
	AccessTokenType op.AccessTokenType                      `json:"access_token_type"`
}

func NewOAuthProvider(application *Application) OAuthProvider {
	return OAuthProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("prv", 12),
		},

		ApplicationID: application.ID,

		RotateRefreshToken: true,
		ApplicationType:    AppTypeConfidential,

		ClientID: crypto.CreateRandString(24),
		Secret:   crypto.CreateRandString(32),

		RedirectURIs:     datatypes.NewJSONType([]string{}),
		RedirectURIGlobs: datatypes.NewJSONType([]string{}),

		ResponseTypes:   datatypes.NewJSONType([]oidc.ResponseType{oidc.ResponseTypeCode, oidc.ResponseTypeIDTokenOnly, oidc.ResponseTypeIDToken}),
		GrantTypes:      datatypes.NewJSONType([]oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken, oidc.GrantTypeTokenExchange}),
		AccessTokenType: op.AccessTokenTypeBearer,
	}
}

func (p *OAuthProvider) RotateSecret() {
	p.Secret = crypto.CreateRandString(32)
}
