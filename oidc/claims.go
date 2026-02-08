package oidc

type Claim string

const (
	// Includes the 'email' and 'email_verified' claims
	ClaimEmail Claim = "email"
	// Includes the 'groups' claim (list of role names)
	ClaimGroups Claim = "groups"
	// Includes the profile claims (name, picture, etc.)
	ClaimProfile Claim = "profile"
	// Includes the 'roles' claim (list of role data)
	ClaimRoles Claim = "star:roles"
	// Includes the 'picture' claim (specified profile picture, or the generated default)
	ClaimPicture Claim = "star:picture"
	// Includes the 'meta' claim (the user's public metadata in star)
	ClaimMetadata Claim = "star:metadata"
)

// Supported claims by Star
var SupportedClaims []Claim = []Claim{
	ClaimEmail, ClaimGroups, ClaimRoles, ClaimProfile, ClaimPicture, ClaimMetadata,
}

// Additional claims that can be requested for userinfo
func CreateClaimsQuery(userInfoClaims []string, idTokenClaims []string) map[string]any {
	values := make(map[string]any)

	if len(userInfoClaims) > 0 {
		claims := make(map[string]any)

		for _, claim := range userInfoClaims {
			claims[claim] = nil
		}

		values["userinfo"] = claims
	}

	if len(idTokenClaims) > 0 {
		claims := make(map[string]any)

		for _, claim := range idTokenClaims {
			claims[claim] = nil
		}

		values["id_token"] = claims
	}

	return values
}
