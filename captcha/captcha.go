package captcha

type Verifier interface {
	GetSiteKey() string
	ValidateToken(token string) bool
}