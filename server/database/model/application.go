package model

import (
	"github.com/tubblelive/star/server/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/datatypes"
)

type Application struct {
	BaseModel
	Secret           string `json:"-"`
	Internal         bool
	RedirectURIs     []ClientRedirectURI `gorm:"many2many:client_redirect_uris"`
	ApplicationType  op.ApplicationType
	AuthMethod       oidc.AuthMethod
	ResponseTypes    datatypes.JSONType[[]oidc.ResponseType]
	GrantTypes       datatypes.JSONType[[]oidc.GrantType]
	AccessTokenType  op.AccessTokenType
	RedirectURIGlobs []ClientPostLogoutRedirectURI `gorm:"many2many:client_post_logout_uris"`
}

type ClientRedirectURI struct {
	BaseModel
	Application   *Application
	ApplicationID string
	URI           string
}

type ClientPostLogoutRedirectURI struct {
	BaseModel
	Application   *Application
	ApplicationID string
	URI           string
}

func NewApplication(name string) Application {
	return Application{
		BaseModel: BaseModel{
			ID: CreateTypedId("app", 12),
		},
		Secret:           crypto.CreateRandString(24),
		Internal:         false,
		RedirectURIs:     []ClientRedirectURI{},
		ApplicationType:  op.ApplicationTypeWeb,
		AuthMethod:       oidc.AuthMethodBasic,
		ResponseTypes:    datatypes.NewJSONType([]oidc.ResponseType{oidc.ResponseTypeCode, oidc.ResponseTypeIDTokenOnly, oidc.ResponseTypeIDToken}),
		GrantTypes:       datatypes.NewJSONType([]oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken, oidc.GrantTypeTokenExchange}),
		AccessTokenType:  op.AccessTokenTypeBearer,
		RedirectURIGlobs: []ClientPostLogoutRedirectURI{},
	}
}
