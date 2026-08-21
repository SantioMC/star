package openid

import (
	"context"
	"strings"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/picture"
)

const RolesClaim = "urn:star:iam:roles"

func UserClaims(
	ctx context.Context,
	db *database.Database,
	user *database.User,
	scopes []string,
) map[string]any {
	claims := map[string]any{
		"sub": user.ID,
	}

	wantsProfile := HasScope(scopes, ScopeUserRead)
	wantsEmail := HasScope(scopes, ScopeUserEmailRead)

	address := ""
	if wantsProfile || wantsEmail {
		if email := db.GetPrimaryEmail(ctx, user.ID); email != nil {
			address = email.Email

			if wantsEmail {
				claims["email"] = email.Email
				claims["email_verified"] = email.Verified
			}
		}
	}

	if wantsProfile {
		claims["name"] = displayName(user)
		claims["nickname"] = user.Name
		claims["preferred_username"] = user.Name

		if firstName := trimmed(user.FirstName); firstName != "" {
			claims["given_name"] = firstName
		}

		if lastName := trimmed(user.LastName); lastName != "" {
			claims["family_name"] = lastName
		}

		if profilePicture := picture.GetProfilePicture(user, address); profilePicture != "" {
			claims["picture"] = profilePicture
		}
	}

	if HasScope(scopes, ScopeUserRolesRead) {
		roles := db.GetUserRoles(ctx, user.ID)
		names := make([]string, len(roles))

		for index, role := range roles {
			names[index] = role.Name
		}

		claims[RolesClaim] = names
	}

	return claims
}

func displayName(user *database.User) string {
	parts := []string{}

	if firstName := trimmed(user.FirstName); firstName != "" {
		parts = append(parts, firstName)
	}

	if lastName := trimmed(user.LastName); lastName != "" {
		parts = append(parts, lastName)
	}

	if len(parts) == 0 {
		return user.Name
	}

	return strings.Join(parts, " ")
}

func trimmed(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}
