package oauth

type Oauth2Error string

const (
	Oauth2ErrorInvalidRequest          Oauth2Error = "invalid_request"
	Oauth2ErrorUnauthorizedClient      Oauth2Error = "unauthorized_client"
	Oauth2ErrorAccessDenied            Oauth2Error = "access_denied"
	Oauth2ErrorUnsupportedResponseType Oauth2Error = "unsupported_response_type"
	Oauth2ErrorInvalidScope            Oauth2Error = "invalid_scope"
	Oauth2ErrorServerError             Oauth2Error = "server_error"
	Oauth2ErrorTemporarilyUnavailable  Oauth2Error = "temporarily_unavailable"
)

type errorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
	State       string `json:"state,omitempty"`
}

func NewError(kind Oauth2Error, message string, state string) errorResponse {
	return errorResponse{
		Error:       string(kind),
		Description: message,
		State:       state,
	}
}
