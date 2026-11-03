package emails

import "net/url"

func VerificationURL(serverURL string, email string, code string) string {
	values := url.Values{}
	values.Set("email", email)
	values.Set("code", code)

	return serverURL + "/verify?" + values.Encode()
}

func PasswordResetURL(serverURL string, token string) string {
	values := url.Values{}
	values.Set("token", token)

	return serverURL + "/reset-password?" + values.Encode()
}

func RecoveryURL(serverURL string) string {
	return serverURL + "/forgot-password"
}
