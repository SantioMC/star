package model

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/tubblelive/star/oidc"
	"gorm.io/datatypes"
)

type SocialProvider struct {
	BaseModel

	Name string `gorm:"unique" json:"name"`

	ClientID     string                       `json:"client_id"`
	ClientSecret string                       `json:"client_secret" encrypt:"true"`
	Scopes       datatypes.JSONType[[]string] `json:"scopes"`

	DiscoveryURL *string `json:"oidc_url"`
	AuthURL      string  `json:"auth_url"`
	TokenURL     string  `json:"token_url"`
}

type LoginState struct {
	BaseModel

	ApplicationID string
	Application   *Application
}

// Reaches out to the oidc discovery endpoint to find the relevant information
func NewSocialProviderFromDiscovery(name string, discoveryUri string, clientId string, clientSecret string) (*SocialProvider, error) {
	config, err := oidc.Discover(discoveryUri)
	if err != nil {
		return nil, err
	}

	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},
		Name: name,

		ClientID:     clientId,
		ClientSecret: clientId,

		DiscoveryURL: &discoveryUri,
		AuthURL:      config.AuthEndpoint,
		TokenURL:     config.TokenEndpoint,
	}, nil
}

// Creates a new social provider from a preset
func NewSocialProviderFromPreset(name string, config oidc.Configuration, clientId string, clientSecret string) (*SocialProvider, error) {
	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},
		Name: name,

		ClientID:     clientId,
		ClientSecret: clientId,

		DiscoveryURL: nil,
		AuthURL:      config.AuthorizationURL,
		TokenURL:     config.TokenURL,
		Scopes:       datatypes.NewJSONType(config.DefaultScopes),
	}, nil
}

// Creates a new social provider from direct values
func NewSocialProvider(name string, authUrl string, tokenUrl string, clientId string, clientSecret string) (*SocialProvider, error) {
	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},
		Name: name,

		ClientID:     clientId,
		ClientSecret: clientId,

		DiscoveryURL: nil,
		AuthURL:      authUrl,
		TokenURL:     tokenUrl,
	}, nil
}

func (p *SocialProvider) GetCallbackURL(hostname string, state string) string {
	return fmt.Sprintf("%s/api/callback/%s", hostname, p.Name)
}

func (p *SocialProvider) GetLoginURL(state string, redirectUri string) string {
	builder := strings.Builder{}

	builder.WriteString(p.AuthURL)
	builder.WriteString("?response_type=code") // todo: We will probably want to support more
	
	builder.WriteString("&client_id=")
	builder.WriteString(url.QueryEscape(p.ClientID))

	builder.WriteString("&scope=")
	builder.WriteString(url.QueryEscape(strings.Join(p.Scopes.Data(), " ")))

	builder.WriteString("&state=")
	builder.WriteString(url.QueryEscape(state))

	builder.WriteString("&redirect_uri=")
	builder.WriteString(url.QueryEscape(redirectUri))

	builder.WriteString("&prompt=consent") // We might want to support others, not sure

	return builder.String()
}
