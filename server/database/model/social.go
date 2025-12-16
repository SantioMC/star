package model

import (
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

func (p *SocialProvider) GetLoginURL() string {
	// ?response_type=code&client_id=157730590492196864&scope=identify%20guilds.join&state=15773059ghq9183habn&redirect_uri=https%3A%2F%2Fnicememe.website&prompt=consent&integration_type=0

	url := p.AuthURL
	url = url + "?response_type=code"
	url = url + "&client_id=" + p.ClientID
	url = url + "&scope=" + p.ClientID

	return url
}
