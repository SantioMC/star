package claims

type UserJWT struct {
	Subject string `json:"sub"`
	Email string `json:"email"`
	Roles []Role `json:"roles"`
	IssuedAt int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
	NotBefore int64 `json:"nbf"`
	NotAfter int64 `json:"naf"`
}

type Role struct {
	ID string `json:"id"`
	Name string `json:"name"`
	SuperRole bool `json:"super"`
}