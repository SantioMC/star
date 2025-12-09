package model

import (
	"log/slog"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/text/language"
	"gorm.io/datatypes"
)

type AuthRequest struct {
	BaseModel
	Code          string
	ApplicationID string
	CallbackURI   string
	TransferState string
	Prompt        datatypes.JSONType[[]string]
	UiLocales     datatypes.JSONType[[]language.Tag]
	LoginHint     string
	MaxAuthAge    *time.Duration
	UserID        string
	Scopes        datatypes.JSONType[[]string]
	ResponseType  oidc.ResponseType
	ResponseMode  oidc.ResponseMode
	Nonce         string
	CodeChallenge *OIDCCodeChallenge `gorm:"embedded"`

	Completed bool
	AuthTime  time.Time
}

type OIDCCodeChallenge struct {
	Challenge string
	Method    string
}

func (a *AuthRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", a.ID),
		slog.Int64("creation_date", a.CreatedAt),
		slog.Any("scopes", a.Scopes),
		slog.String("response_type", string(a.ResponseType)),
		slog.String("app_id", a.ApplicationID),
		slog.String("callback_uri", a.CallbackURI),
	)
}

func (a *AuthRequest) GetID() string {
	return a.ID
}

func (a *AuthRequest) GetACR() string {
	return ""
}

func (a *AuthRequest) GetAMR() []string {
	if a.Done() {
		return []string{"pwd"}
	}

	return nil
}

func (a *AuthRequest) GetAudience() []string {
	return []string{a.ApplicationID}
}

func (a *AuthRequest) GetAuthTime() time.Time {
	return a.AuthTime
}

func (a *AuthRequest) GetClientID() string {
	return a.ApplicationID
}

func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	panic("unimplemented")
	// return CodeChallengeToOIDC(a.CodeChallenge)
}

func (a *AuthRequest) GetNonce() string {
	return a.Nonce
}

func (a *AuthRequest) GetRedirectURI() string {
	return a.CallbackURI
}

func (a *AuthRequest) GetResponseType() oidc.ResponseType {
	return a.ResponseType
}

func (a *AuthRequest) GetResponseMode() oidc.ResponseMode {
	return a.ResponseMode
}

func (a *AuthRequest) GetScopes() []string {
	return a.Scopes.Data()
}

func (a *AuthRequest) GetState() string {
	return a.TransferState
}

func (a *AuthRequest) GetSubject() string {
	return a.UserID
}

func (a *AuthRequest) Done() bool {
	return a.Completed
}