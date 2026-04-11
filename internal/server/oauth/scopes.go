package oauth

import "slices"

// Common / Default scopes
var ScopeOpenID = "openid"
var ScopeProfile = "profile"
var ScopeEmail = "email"
var ScopeOfflineAccess = "offline_access"

var SupportedScopes = []string{
	ScopeOpenID, ScopeProfile, ScopeEmail, ScopeOfflineAccess,
}

func ValidateScopes(scopes []string) ([]string, bool) {
	someInvalid := false
	var validScopes []string

	for _, scope := range scopes {
		if scope == "" {
			continue
		}

		if slices.Contains(SupportedScopes, scope) {
			validScopes = append(validScopes, scope)
		} else {
			someInvalid = true
		}
	}

	return validScopes, someInvalid
}
