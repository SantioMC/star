package captcha

type Verifier interface {
	ValidateToken(token string) bool
}