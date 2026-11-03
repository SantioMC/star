package federation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/tubblelive/star/internal/database"
)

type Tokens struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Authorization struct {
	RedirectURI string
	State       string
	Nonce       string
	Challenge   string
}

func AuthorizationURL(endpoints *Endpoints, provider *database.Provider, authorization *Authorization) (string, error) {
	target, err := url.Parse(endpoints.AuthorizationURL)
	if err != nil {
		return "", err
	}

	query := target.Query()
	query.Set("response_type", "code")
	query.Set("client_id", provider.ClientID)
	query.Set("redirect_uri", authorization.RedirectURI)
	query.Set("scope", strings.Join(provider.Scopes, " "))
	query.Set("state", authorization.State)
	query.Set("code_challenge", authorization.Challenge)
	query.Set("code_challenge_method", "S256")

	if slices.Contains(provider.Scopes, "openid") {
		query.Set("nonce", authorization.Nonce)
	}

	target.RawQuery = query.Encode()
	return target.String(), nil
}

func Exchange(
	ctx context.Context,
	endpoints *Endpoints,
	provider *database.Provider,
	code string,
	redirectURI string,
	verifier string,
) (*Tokens, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", provider.ClientID)
	form.Set("client_secret", provider.ClientSecret)
	form.Set("code_verifier", verifier)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoints.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var tokens Tokens
	if err := doJSON(request, &tokens); err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if tokens.Error != "" {
		return nil, fmt.Errorf("provider rejected the authorization code: %s %s", tokens.Error, tokens.ErrorDescription)
	}

	if tokens.AccessToken == "" {
		return nil, errors.New("provider did not return an access token")
	}

	return &tokens, nil
}
