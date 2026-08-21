package openid

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/tubblelive/star/internal/database"
)

const StarNamespace = "star"
const starScopePrefix = StarNamespace + ":"

const (
	ScopeOpenID        = "openid"
	ScopeOfflineAccess = "offline_access"

	ScopeProfile = "profile"
	ScopeOIDC    = "oidc"
	ScopeEmail   = "email"
	ScopeRoles   = "roles"

	ScopeUserRead          = "star:user:read"
	ScopeUserWrite         = "star:user:write"
	ScopeUserPasswordWrite = "star:user:password:write"
	ScopeUserEmailRead     = "star:user:email:read"
	ScopeUserEmailWrite    = "star:user:email:write"
	ScopeUserRolesRead     = "star:user:roles:read"
	ScopeUserSessionsRead  = "star:user:sessions:read"
	ScopeUserSessionsWrite = "star:user:sessions:write"
	ScopeUserMFARead       = "star:user:mfa:read"
	ScopeUserMFAWrite      = "star:user:mfa:write"
	ScopeUserAuthorizeApps = "star:user:applications:authorize"

	ScopeApplicationsRead   = "star:applications:read"
	ScopeApplicationsWrite  = "star:applications:write"
	ScopeApplicationsDelete = "star:applications:delete"

	ScopeAdminUsersRead      = "star:admin:users:read"
	ScopeAdminUsersWrite     = "star:admin:users:write"
	ScopeAdminUsersDelete    = "star:admin:users:delete"
	ScopeAdminRolesRead      = "star:admin:roles:read"
	ScopeAdminRolesWrite     = "star:admin:roles:write"
	ScopeAdminProvidersRead  = "star:admin:providers:read"
	ScopeAdminProvidersWrite = "star:admin:providers:write"
	ScopeAdminSettingsRead   = "star:admin:settings:read"
	ScopeAdminSettingsWrite  = "star:admin:settings:write"
)

type ScopeDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases,omitempty"`
	Restricted  bool     `json:"restricted"`
}

var BuiltinScopes = []ScopeDefinition{
	{
		Name:        ScopeOpenID,
		Description: "Confirm your identity and issue an identity token",
	},
	{
		Name:        ScopeOfflineAccess,
		Description: "Stay signed in and renew access without asking you again",
	},
	{
		Name:        ScopeUserRead,
		Description: "Read your basic profile information",
		Aliases:     []string{ScopeProfile, ScopeOIDC},
	},
	{
		Name:        ScopeUserWrite,
		Description: "Update your username and display name",
	},
	{
		Name:        ScopeUserEmailRead,
		Description: "Read your email addresses and whether they are verified",
		Aliases:     []string{ScopeEmail},
	},
	{
		Name:        ScopeUserEmailWrite,
		Description: "Add, remove, and change the primary of your email addresses",
	},
	{
		Name:        ScopeUserRolesRead,
		Description: "Read the roles assigned to your account",
		Aliases:     []string{ScopeRoles},
	},
	{
		Name:        ScopeUserSessionsRead,
		Description: "View the devices and sessions signed in to your account",
	},
	{
		Name:        ScopeUserSessionsWrite,
		Description: "Sign your account out of its active sessions",
	},
	{
		Name:        ScopeUserMFARead,
		Description: "View which multi-factor methods your account has enrolled",
	},
	{
		Name:        ScopeUserMFAWrite,
		Description: "Enroll or remove multi-factor methods on your account",
		Restricted:  true,
	},
	{
		Name:        ScopeUserPasswordWrite,
		Description: "Change or remove your account password",
		Restricted:  true,
	},
	{
		Name:        ScopeUserAuthorizeApps,
		Description: "Authorize other applications to access your account",
		Restricted:  true,
	},
	{
		Name:        ScopeApplicationsRead,
		Description: "View applications registered in star",
		Restricted:  true,
	},
	{
		Name:        ScopeApplicationsWrite,
		Description: "Create and modify applications registered in star",
		Restricted:  true,
	},
	{
		Name:        ScopeApplicationsDelete,
		Description: "Delete applications registered in star",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminUsersRead,
		Description: "View every user within the server",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminUsersWrite,
		Description: "Modify every user within the server",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminUsersDelete,
		Description: "Delete users and their sessions",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminRolesRead,
		Description: "View roles and their permissions",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminRolesWrite,
		Description: "Create, modify, and delete roles",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminProvidersRead,
		Description: "View configured social login providers",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminProvidersWrite,
		Description: "Create, modify, and delete social login providers",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminSettingsRead,
		Description: "View the server configuration",
		Restricted:  true,
	},
	{
		Name:        ScopeAdminSettingsWrite,
		Description: "Modify the server configuration",
		Restricted:  true,
	},
}

var builtinIndex = func() map[string]*ScopeDefinition {
	index := make(map[string]*ScopeDefinition)

	for idx := range BuiltinScopes {
		definition := &BuiltinScopes[idx]
		index[definition.Name] = definition

		for _, alias := range definition.Aliases {
			index[alias] = definition
		}
	}

	return index
}()

var ReservedNamespaces = []string{
	StarNamespace, ScopeOpenID, ScopeOIDC, ScopeProfile,
	ScopeEmail, ScopeRoles, ScopeOfflineAccess,
}

var namespacePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,31}$`)
var namespaceSanitizer = regexp.MustCompile(`[^a-z0-9]+`)
var scopeKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*(:[a-z0-9][a-z0-9_.-]*)*$`)

func ValidNamespace(namespace string) bool {
	return namespacePattern.MatchString(namespace)
}

func NormalizeNamespace(value string) string {
	lowered := strings.ToLower(strings.TrimSpace(value))
	sanitized := namespaceSanitizer.ReplaceAllString(lowered, "-")
	trimmed := strings.Trim(sanitized, "-")

	if len(trimmed) > 32 {
		trimmed = strings.Trim(trimmed[:32], "-")
	}

	return trimmed
}

func ReservedNamespace(namespace string) bool {
	return slices.Contains(ReservedNamespaces, namespace)
}

func ValidScopeName(namespace string, scope string) bool {
	if len(scope) > 128 {
		return false
	}

	prefix := namespace + ":"
	if !strings.HasPrefix(scope, prefix) {
		return false
	}

	key := strings.TrimPrefix(scope, prefix)
	if key == "" {
		return false
	}

	return scopeKeyPattern.MatchString(key)
}

func Builtin(scope string) *ScopeDefinition {
	return builtinIndex[scope]
}

func Canonical(scope string) string {
	if definition := Builtin(scope); definition != nil {
		return definition.Name
	}

	return scope
}

func IsStarScope(scope string) bool {
	return strings.HasPrefix(Canonical(scope), starScopePrefix)
}

func HasScope(granted []string, required string) bool {
	for _, scope := range granted {
		if Canonical(scope) == required {
			return true
		}
	}

	return false
}

func SupportedScopes() []string {
	names := make([]string, 0, len(BuiltinScopes))

	for _, definition := range BuiltinScopes {
		if definition.Restricted {
			continue
		}

		names = append(names, definition.Name)
		names = append(names, definition.Aliases...)
	}

	return names
}

func EffectiveScopes(granted []string, firstParty bool) []string {
	if !firstParty || len(granted) > 0 {
		return granted
	}

	names := make([]string, 0, len(BuiltinScopes))
	for _, definition := range BuiltinScopes {
		names = append(names, definition.Name)
	}

	return names
}

type ResolvedScope struct {
	Scope           string
	Canonical       string
	Description     string
	ApplicationID   string
	ApplicationName string
	Restricted      bool
	Requestable     bool
}

type ScopeResolution struct {
	Granted []ResolvedScope
	Invalid []string
}

func ResolveScopes(
	ctx context.Context,
	db *database.Database,
	application *database.Application,
	requested []string,
) ScopeResolution {
	resolution := ScopeResolution{
		Granted: []ResolvedScope{},
		Invalid: []string{},
	}

	seen := make(map[string]bool)
	ordered := make([]string, 0, len(requested))
	unknown := make([]string, 0, len(requested))

	for _, scope := range requested {
		if scope == "" || seen[scope] {
			continue
		}

		seen[scope] = true
		ordered = append(ordered, scope)

		if Builtin(scope) == nil {
			unknown = append(unknown, scope)
		}
	}

	registered := make(map[string]database.OwnedApplicationScope)
	for _, scope := range db.GetApplicationScopesByNames(ctx, unknown) {
		registered[scope.Name] = scope
	}

	internal := application != nil && application.Internal

	for _, scope := range ordered {
		if definition := Builtin(scope); definition != nil {
			if definition.Restricted && !internal {
				resolution.Invalid = append(resolution.Invalid, scope)
				continue
			}

			resolution.Granted = append(resolution.Granted, ResolvedScope{
				Scope:       scope,
				Canonical:   definition.Name,
				Description: definition.Description,
				Restricted:  definition.Restricted,
			})

			continue
		}

		owned, found := registered[scope]
		if !found {
			resolution.Invalid = append(resolution.Invalid, scope)
			continue
		}

		owner := application != nil && application.ID == owned.ApplicationID
		if !owned.Requestable && !owner {
			resolution.Invalid = append(resolution.Invalid, scope)
			continue
		}

		resolution.Granted = append(resolution.Granted, ResolvedScope{
			Scope:           scope,
			Canonical:       scope,
			Description:     owned.Description,
			ApplicationID:   owned.ApplicationID,
			ApplicationName: owned.ApplicationName,
			Requestable:     owned.Requestable,
		})
	}

	return resolution
}

func (resolution ScopeResolution) Names() []string {
	names := make([]string, len(resolution.Granted))

	for idx, scope := range resolution.Granted {
		names[idx] = scope.Scope
	}

	return names
}

func (resolution ScopeResolution) Audiences(applicationID string, defaultApplicationID string) []string {
	audiences := []string{applicationID}

	for _, scope := range resolution.Granted {
		audience := scope.ApplicationID

		if audience == "" {
			if !strings.HasPrefix(scope.Canonical, starScopePrefix) {
				continue
			}

			audience = defaultApplicationID
		}

		if !slices.Contains(audiences, audience) {
			audiences = append(audiences, audience)
		}
	}

	return audiences
}
