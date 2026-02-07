package oidc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var ErrRequestFailed = fmt.Errorf("request returned a non-200 status code")

type OpenIDDiscovery struct {
	Issuer           string  `json:"issuer"`
	AuthEndpoint     string  `json:"authorization_endpoint"`
	TokenEndpoint    string  `json:"token_endpoint"`

	RegisterEndpoint *string `json:"registration_endpoint"`
	UserInfoEndpoint *string `json:"userinfo_endpoint"`
	JwksEndpoint     *string `json:"jwks_uri"`

	SupportedTokenAuth *[]string `json:"token_endpoint_auth_methods_supported"`
	SupportedTokenAlg  *[]string `json:"token_endpoint_auth_signing_alg_values_supported"`

	SupportedScopes *[]string `json:"scopes_supported"`
}

type TokenExchange struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserInfo struct {
	ID             string `json:"sub"`
	Username       string `json:"preferred_username"`
	FirstName      string `json:"given_name"`
	LastName       string `json:"family_name"`
	ProfilePicture string `json:"picture"`
	Email          string `json:"email"`
}

func Discover(discoveryUri string) (*OpenIDDiscovery, error) {
	res, err := http.Get(discoveryUri)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, ErrRequestFailed
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data *OpenIDDiscovery
	err = json.Unmarshal(body, &body)

	if err != nil {
		return nil, err
	}

	return data, nil
}
