package oidc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var ErrRequestFailed = fmt.Errorf("discovery url returned a non-200 status code")

// Not a complete model of the /.well-known/openid-configuration, but contains only the info we care about
type OpenIDConfiguration struct {
	Issuer           string `json:"issuer"`
	AuthEndpoint     string `json:"authorization_endpoint"`
	TokenEndpoint    string `json:"token_endpoint"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
	JwksEndpoint     string `json:"jwks_uri"`
}

func Discover(discoveryUri string) (*OpenIDConfiguration, error) {
	res, err := http.Get(discoveryUri)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if (res.StatusCode != 200) {
		return nil, ErrRequestFailed
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data *OpenIDConfiguration
	err = json.Unmarshal(body, &body)

	if err != nil {
		return nil, err
	}

	return data, nil
}