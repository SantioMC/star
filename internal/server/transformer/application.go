package transformer

import (
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
)

type AdminApplication struct {
	database.Application
	HasClientSecret bool                        `json:"has_client_secret"`
	Scopes          []database.ApplicationScope `json:"scopes"`
	Roles           []string                    `json:"roles"`
	BuiltinScopes   []openid.ScopeDefinition    `json:"builtin_scopes,omitempty"`
}

type SharedApplication struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Authorized bool   `json:"authorized"`
}

func TransformSharedApplications(applications []database.SharedApplication) []SharedApplication {
	transformed := make([]SharedApplication, len(applications))

	for idx, application := range applications {
		transformed[idx] = SharedApplication{
			ID:         application.ID,
			Name:       application.Name,
			URL:        application.ApplicationURL,
			Authorized: application.Authorized,
		}
	}

	return transformed
}

func builtinScopes(application database.Application) []openid.ScopeDefinition {
	if !application.Internal || application.Namespace != openid.StarNamespace {
		return nil
	}

	return openid.BuiltinScopes
}

func TransformApplication(application database.Application, scopes []database.ApplicationScope, roles []string) AdminApplication {
	if scopes == nil {
		scopes = []database.ApplicationScope{}
	}

	if roles == nil {
		roles = []string{}
	}

	return AdminApplication{
		Application:     application,
		HasClientSecret: application.ClientSecret != nil,
		Scopes:          scopes,
		Roles:           roles,
		BuiltinScopes:   builtinScopes(application),
	}
}

func TransformApplications(
	applications []database.Application,
	scopes []database.ApplicationScope,
	bindings []database.ApplicationRoleBinding,
) []AdminApplication {
	grouped := make(map[string][]database.ApplicationScope, len(applications))
	for _, scope := range scopes {
		grouped[scope.ApplicationID] = append(grouped[scope.ApplicationID], scope)
	}

	roles := make(map[string][]string, len(applications))
	for _, binding := range bindings {
		roles[binding.ApplicationID] = append(roles[binding.ApplicationID], binding.RoleID)
	}

	transformed := make([]AdminApplication, len(applications))
	for idx, application := range applications {
		transformed[idx] = TransformApplication(application, grouped[application.ID], roles[application.ID])
	}

	return transformed
}
