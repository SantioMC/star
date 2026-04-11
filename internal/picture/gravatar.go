package picture

import (
	"crypto/sha256"
	"fmt"

	"github.com/tubblelive/star/internal/database"
)

func GetGravatarProfilePicture(user *database.User, email string) string {
	hash := sha256.Sum256([]byte(email))
	initials := ""

	if user.FirstName != nil {
		value := *user.FirstName
		if len(value) > 1 {
			initials += value[:1]
		}
	}

	if user.LastName != nil {
		value := *user.LastName
		if len(value) > 1 {
			initials += value[:1]
		}
	}

	if initials == "" {
		initials = user.Name[:2]
	}

	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=initials&initials=%s", hash, initials)
}
