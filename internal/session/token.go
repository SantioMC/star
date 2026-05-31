package session

import (
	"fmt"

	"github.com/tubblelive/star/internal/cryptography"
)

func CreateSessionToken() string {
	return fmt.Sprintf("star_%s", cryptography.CreateSecureString(32))
}
