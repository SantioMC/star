package oauth

import (
	"slices"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
)

func canonicalScopes(scopes []string) []string {
	canonical := make([]string, len(scopes))

	for idx, scope := range scopes {
		canonical[idx] = openid.Canonical(scope)
	}

	return canonical
}

func RequiresConsent(application *database.Application, grant *database.ApplicationGrant, scopes []string) bool {
	switch application.Consent {
	case database.ApplicationConsentImplicit:
		return false
	case database.ApplicationConsentExplicit:
		return true
	}

	if grant == nil {
		return true
	}

	for _, scope := range canonicalScopes(scopes) {
		if !slices.Contains(grant.Scopes, scope) {
			return true
		}
	}

	return false
}

func MergeGrantScopes(grant *database.ApplicationGrant, scopes []string) []string {
	merged := canonicalScopes(scopes)
	if grant != nil {
		merged = append(merged, grant.Scopes...)
	}

	slices.Sort(merged)
	return slices.Compact(merged)
}
