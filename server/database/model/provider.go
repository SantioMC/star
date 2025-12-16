package model

import (
	"github.com/tubblelive/star/server/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/datatypes"
)

type ProviderKind int

const (
	OAuth2 ProviderKind = iota
)

type Provider struct {
	BaseModel

	Application   *Application `json:"-"`
	ApplicationID string       `json:"application_id"`
	Type          ProviderKind `json:"type"`

	Secret           string `json:"secret" encrypt:"true"`
	RedirectURIs     datatypes.JSONType[[]string]
	RedirectURIGlobs datatypes.JSONType[[]string]

	ApplicationType op.ApplicationType
	AuthMethod      oidc.AuthMethod
	ResponseTypes   datatypes.JSONType[[]oidc.ResponseType]
	GrantTypes      datatypes.JSONType[[]oidc.GrantType]
	AccessTokenType op.AccessTokenType
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
