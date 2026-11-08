package captcha

type NoopVerifier struct{}

func (v *NoopVerifier) GetSiteKey() string {
	return ""
}

func (v *NoopVerifier) ValidateToken(token string) bool {
	return true
}

func (v *NoopVerifier) Check(token string) error {
	return nil
}
