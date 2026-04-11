package session

import "github.com/tubblelive/star/internal/cryptography"

func CreateSessionToken() string {
	return cryptography.CreateSecureString(32)
}
