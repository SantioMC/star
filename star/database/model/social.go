package model

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/tubblelive/star/oidc"
	"gorm.io/datatypes"
)

var ErrRequestFailed = fmt.Errorf("response returned a non-200 status code")
var ErrNoUserinfoEndpoint = fmt.Errorf("no oidc userinfo field was found")

type SocialProvider struct {
	BaseModel

	Name string `gorm:"unique" json:"name"`

	ClientID     string                       `json:"client_id"`
	ClientSecret string                       `json:"client_secret" encrypt:"true"`
	Scopes       datatypes.JSONType[[]string] `json:"scopes"`
	Icon         string                       `json:"icon"`

	DiscoveryURL *string `json:"oidc_url"`
	AuthURL      string  `json:"auth_url"`
	TokenURL     string  `json:"token_url"`
	UserInfoURL  string  `json:"userinfo_url"`

	IDField       string `json:"id_field"`
	UsernameField string `json:"username_field"`
	EmailField    string `json:"email_field"`
	AvatarField   string `json:"avatar_field"`
}

type LoginState struct {
	State     string `gorm:"primaryKey"`
	CreatedAt int64  `gorm:"autoCreateTime"`

	ApplicationID string
	Application   *Application

	UserAgent string
	Address   string
}

// Reaches out to the oidc discovery endpoint to find the relevant information
func NewSocialProviderFromDiscovery(name string, discoveryUri string, clientId string, clientSecret string) (*SocialProvider, error) {
	config, err := oidc.Discover(discoveryUri)
	if err != nil {
		return nil, err
	}

	if config.UserInfoEndpoint == nil {
		return nil, ErrNoUserinfoEndpoint
	}

	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},
		Name: name,

		ClientID:     clientId,
		ClientSecret: clientSecret,

		DiscoveryURL: &discoveryUri,
		AuthURL:      config.AuthEndpoint,
		TokenURL:     config.TokenEndpoint,
		UserInfoURL:  *config.UserInfoEndpoint,
	}, nil
}

// Creates a new social provider from a preset
func NewSocialProviderFromPreset(name string, config oidc.Configuration, clientId string, clientSecret string) (*SocialProvider, error) {
	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},

		Name: name,
		Icon: config.Icon,

		ClientID:     clientId,
		ClientSecret: clientSecret,

		DiscoveryURL: nil,
		AuthURL:      config.AuthorizationURL,
		TokenURL:     config.TokenURL,
		UserInfoURL:  config.UserInfoURL,

		Scopes: datatypes.NewJSONType(config.DefaultScopes),
	}, nil
}

// Creates a new empty social provider from direct values
func NewSocialProvider(name string, clientId string, clientSecret string) (*SocialProvider, error) {
	return &SocialProvider{
		BaseModel: BaseModel{
			ID: CreateTypedId("scp", 12),
		},
		Name: name,

		ClientID:     clientId,
		ClientSecret: clientSecret,

		DiscoveryURL: nil,
	}, nil
}

func (p *SocialProvider) IsValid() bool {
	return p.AuthURL != "" && p.TokenURL != "" && p.UserInfoURL != ""
}

func (p *SocialProvider) GetCallbackURL(hostname string, state string) string {
	return fmt.Sprintf("%s/api/callback/%s", hostname, p.Name)
}

func (p *SocialProvider) GetLoginURL(state string, redirectUri string) (string, error) {
	builder := strings.Builder{}

	claims := oidc.CreateClaimsQuery([]string{"email", "picture", "preferred_username"}, []string{})
	encoded, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

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

	builder.WriteString("&claims=")
	builder.WriteString(url.QueryEscape(string(encoded)))

	builder.WriteString("&prompt=consent") // We might want to support others, not sure

	return builder.String(), nil
}

func (p *SocialProvider) ExchangeCode(hostname string, code string) (*oidc.TokenExchange, error) {
	redirectUri := p.GetCallbackURL(hostname, p.Name)

	// We'll just include client details in both body and header to support
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectUri)
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)

	request, err := http.NewRequest("POST", p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(url.QueryEscape(p.ClientID), url.QueryEscape(p.ClientSecret))

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		slog.Error("failed to exchange code for access token", "body", string(body))
		return nil, ErrRequestFailed
	}

	var data oidc.TokenExchange
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil

}

func (p *SocialProvider) GetUserInfo(token string) (*oidc.UserInfo, error) {
	if p.UserInfoURL == "https://discord.com/api/oauth2/userinfo" {
		// Discord at the time of writing has their userinfo endpoint not working with auth, so we'll use their /users/@me endpoint
		return oidc.GetDiscordUserInfo(token)
	}

	req, err := http.NewRequest("GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		slog.Error("failed to get user info", "body", string(body))
		return nil, ErrRequestFailed
	}

	var data oidc.UserInfo
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
