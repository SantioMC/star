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
