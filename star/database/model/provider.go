package model

import (
	"github.com/tubblelive/star/star/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/datatypes"
)

type ProviderKind string

const (
	OAuth2 ProviderKind = "oauth2"
)

var AllProviderKinds = []ProviderKind{OAuth2}

type Provider struct {
	BaseModel

	Application   *Application `json:"-"`
	ApplicationID string       `json:"application_id"`
	Type          ProviderKind `json:"type"`

	Secret           string                       `json:"secret" encrypt:"true"`
	RedirectURIs     datatypes.JSONType[[]string] `json:"redirect_uris"`
	RedirectURIGlobs datatypes.JSONType[[]string] `json:"redirect_uri_globs"`

	ApplicationType op.ApplicationType                      `json:"application_type"`
	AuthMethod      oidc.AuthMethod                         `json:"auth_method"`
	ResponseTypes   datatypes.JSONType[[]oidc.ResponseType] `json:"response_types"`
	GrantTypes      datatypes.JSONType[[]oidc.GrantType]    `json:"grant_types"`
	AccessTokenType op.AccessTokenType                      `json:"access_token_type"`
}

func NewProvider(application *Application, kind ProviderKind) Provider {
	return Provider{
		BaseModel: BaseModel{
			ID: CreateTypedId("prv", 12),
		},

		ApplicationID: application.ID,
		Type:          kind,

		Secret:           crypto.CreateRandString(24),
		RedirectURIs:     datatypes.NewJSONType([]string{}),
		RedirectURIGlobs: datatypes.NewJSONType([]string{}),

		ApplicationType: op.ApplicationTypeWeb,
		AuthMethod:      oidc.AuthMethodBasic,
		ResponseTypes:   datatypes.NewJSONType([]oidc.ResponseType{oidc.ResponseTypeCode, oidc.ResponseTypeIDTokenOnly, oidc.ResponseTypeIDToken}),
		GrantTypes:      datatypes.NewJSONType([]oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken, oidc.GrantTypeTokenExchange}),
		AccessTokenType: op.AccessTokenTypeBearer,
	}
}

func (p *Provider) RotateSecret() {
	p.Secret = crypto.CreateRandString(24)
}
