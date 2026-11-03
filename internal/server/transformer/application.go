package transformer

import (
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
)

type AdminApplication struct {
	database.Application
	HasClientSecret bool                        `json:"has_client_secret"`
	Scopes          []database.ApplicationScope `json:"scopes"`
	BuiltinScopes   []openid.ScopeDefinition    `json:"builtin_scopes,omitempty"`
}

func builtinScopes(application database.Application) []openid.ScopeDefinition {
	if !application.Internal || application.Namespace != openid.StarNamespace {
		return nil
	}

	return openid.BuiltinScopes
}

func TransformApplication(application database.Application, scopes []database.ApplicationScope) AdminApplication {
	if scopes == nil {
		scopes = []database.ApplicationScope{}
	}

	return AdminApplication{
		Application:     application,
		HasClientSecret: application.ClientSecret != nil,
		Scopes:          scopes,
		BuiltinScopes:   builtinScopes(application),
	}
}

func TransformApplications(applications []database.Application, scopes []database.ApplicationScope) []AdminApplication {
	grouped := make(map[string][]database.ApplicationScope, len(applications))
	for _, scope := range scopes {
		grouped[scope.ApplicationID] = append(grouped[scope.ApplicationID], scope)
	}

	transformed := make([]AdminApplication, len(applications))
	for idx, application := range applications {
		transformed[idx] = TransformApplication(application, grouped[application.ID])
	}

	return transformed
}
