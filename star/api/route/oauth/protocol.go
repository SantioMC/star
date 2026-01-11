package oauth

type TokenErrorCode string
type GrantType string

const (
	TokenErrInvalidRequest       TokenErrorCode = "invalid_request"
	TokenErrInvalidGrant         TokenErrorCode = "invalid_grant"
	TokenErrInvalidClient        TokenErrorCode = "invalid_client"
	TokenErrUnauthorizedClient   TokenErrorCode = "unauthorized_client"
	TokenErrUnsupportedGrantType TokenErrorCode = "unsupported_grant_type"
	TokenErrInvalidScope         TokenErrorCode = "invalid_scope"
	TokenErrServerError          TokenErrorCode = "server_error"

	GrantAuthorizationCode GrantType = "authorization_code"
	GrantRefreshToken      GrantType = "refresh_token"
)

var KnownGrantTypes = []GrantType{
	GrantAuthorizationCode, GrantRefreshToken,
}

type TokenRequest struct {
	GrantType    GrantType `form:"grant_type"`
	Fingerprint  *string   `form:"fp"`
	ClientId     *string   `form:"client_id"`
	ClientSecret *string   `form:"client_secret"`
}

// grant_type = "refresh_token"
type RefreshTokenRequest struct {
	TokenRequest
	RefreshToken string  `form:"refresh_token"`
	Scope        *string `form:"scope"`
}

// grant_type = "authorization_code"
type ExchangeTokenRequest struct {
	TokenRequest
	Code        string  `form:"code"`
	RedirectUri string  `form:"redirect_uri"`
	ClientId    *string `form:"client_id"`
}

type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	Expires      int64   `json:"expires"`
}

type TokenError struct {
	Error       TokenErrorCode `json:"error"`
	Description string         `json:"error_description,omitempty"`
	URI         string         `json:"error_uri,omitempty"`
}
