package storage

import (
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

var(
	clients = map[string]*Client{}
)

type Client struct {
	id                             string
	secret                         string
	redirectURIs                   []string
	applicationType                op.ApplicationType
	authMethod                     oidc.AuthMethod
	responseTypes                  []oidc.ResponseType
	grantTypes                     []oidc.GrantType
	accessTokenType                op.AccessTokenType
	postLogoutRedirectURIGlobs     []string
	redirectURIGlobs               []string
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) RedirectURIs() []string {
	return c.redirectURIs
}

func (c *Client) PostLogoutRedirectURIs() []string {
	return []string{}
}

func (c *Client) ApplicationType() op.ApplicationType {
	return c.applicationType
}

func (c *Client) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

func (c *Client) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}
func (c *Client) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

func (c *Client) LoginURL(id string) string {
	return "/login/" + id
}

func (c *Client) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}

func (c *Client) IDTokenLifetime() time.Duration {
	return 1 * time.Hour
}

func (c *Client) DevMode() bool {
	return false
}

func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *Client) IsScopeAllowed(scope string) bool {
	// TODO: implement
	return false
}

func (c *Client) IDTokenUserinfoClaimsAssertion() bool {
	return true
}

func (c *Client) ClockSkew() time.Duration {
	return time.Second * 10
}

// RegisterClients enables you to register clients for the example implementation
// there are some clients (web and native) to try out different cases
// add more if necessary
//
// RegisterClients should be called before the Storage is used so that there are
// no race conditions.
func RegisterClients(registerClients ...*Client) {
	for _, client := range registerClients {
		clients[client.id] = client
	}
}

// NativeClient will create a client of type native, which will always use PKCE and allow the use of refresh tokens
// user-defined redirectURIs may include:
// - http://localhost without port specification (e.g. http://localhost/auth/callback)
// - custom protocol (e.g. custom://auth/callback)
// (the examples will be used as default, if none is provided)
func NativeClient(id string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost/auth/callback",
			"custom://auth/callback",
		}
	}
	return &Client{
		id:                             id,
		secret:                         "", // no secret needed (due to PKCE)
		redirectURIs:                   redirectURIs,
		applicationType:                op.ApplicationTypeNative,
		authMethod:                     oidc.AuthMethodNone,
		responseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode},
		grantTypes:                     []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		accessTokenType:                op.AccessTokenTypeBearer,
	}
}

// WebClient will create a client of type web, which will always use Basic Auth and allow the use of refresh tokens
// user-defined redirectURIs may include:
// - http://localhost with port specification (e.g. http://localhost:9999/auth/callback)
// (the example will be used as default, if none is provided)
func WebClient(id, secret string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost:9999/auth/callback",
		}
	}
	return &Client{
		id:                             id,
		secret:                         secret,
		redirectURIs:                   redirectURIs,
		applicationType:                op.ApplicationTypeWeb,
		authMethod:                     oidc.AuthMethodBasic,
		responseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode, oidc.ResponseTypeIDTokenOnly, oidc.ResponseTypeIDToken},
		grantTypes:                     []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken, oidc.GrantTypeTokenExchange},
		accessTokenType:                op.AccessTokenTypeBearer,
	}
}

// DeviceClient creates a device client with Basic authentication.
func DeviceClient(id, secret string) *Client {
	return &Client{
		id:                             id,
		secret:                         secret,
		redirectURIs:                   nil,
		applicationType:                op.ApplicationTypeWeb,
		authMethod:                     oidc.AuthMethodBasic,
		responseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode},
		grantTypes:                     []oidc.GrantType{oidc.GrantTypeDeviceCode},
		accessTokenType:                op.AccessTokenTypeBearer,
	}
}

type hasRedirectGlobs struct {
	*Client
}

// RedirectURIGlobs provide wildcarding for additional valid redirects
func (c hasRedirectGlobs) RedirectURIGlobs() []string {
	return c.redirectURIGlobs
}

// PostLogoutRedirectURIGlobs provide extra wildcarding for additional valid redirects
func (c hasRedirectGlobs) PostLogoutRedirectURIGlobs() []string {
	return c.postLogoutRedirectURIGlobs
}

// RedirectGlobsClient wraps the client in a op.HasRedirectGlobs
// only if DevMode is enabled.
func RedirectGlobsClient(client *Client) op.Client {
	return client
}