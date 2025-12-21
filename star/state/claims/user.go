package claims

import "github.com/tubblelive/star/star/state/scopes"

type UserJWT struct {
	Subject   string         `json:"sub"`
	Email     string         `json:"email"`
	Audience  string         `json:"aud"`
	Roles     []Role         `json:"roles"`
	Scopes    []scopes.Scope `json:"scopes"`
	IssuedAt  int64          `json:"iat"`
	ExpiresAt int64          `json:"exp"`
	NotBefore int64          `json:"nbf"`
	NotAfter  int64          `json:"naf"`
}

type Role struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SuperRole bool   `json:"super"`
}
