package picture

import "github.com/tubblelive/star/internal/database"

func GetProfilePicture(user *database.User, email *string) string {
	if user.ProfilePicture != nil {
		return *user.ProfilePicture
	}

	if email == nil || *email == "" {
		return ""
	}

	return GetGravatarProfilePicture(user, *email)
}
