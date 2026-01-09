package scopes

import "strings"

type Scope struct {
	Node string
	Description string
}

func NewScope(node string, description string) Scope {
	return Scope{
		Node: node,
		Description: description,
	}
}

var (
	ScopeAll Scope = NewScope("star:*", "Grants full access to star's management console")
	
	// User scopes, given to all users (starts with `star:user:`)
	ScopeUserAll Scope = NewScope("star:user:*", "Grants full access to user-related permissions")
	ScopeReadEmail Scope = NewScope("star:user:email", "Permits reading a user's email")
	ScopeReadSessions Scope = NewScope("star:user:read:sessions", "Permits reading a user's sessions")
	ScopeDeleteSessions Scope = NewScope("star:user:delete:sessions", "Permits invalidating a user's sessions")
	ScopeDeleteUser Scope = NewScope("star:user:delete", "Permits the deletion of a user's entire account")
	ScopeChangePassword Scope = NewScope("star:user:change_password", "Permits the changing of a user's password")

	// Additional scopes, must be granted
	ScopeReadMetrics Scope = NewScope("star:server:metrics", "Permits viewing star's generic metrics")

	// Generic admin scopes, must be granted
	ScopeAdminAll Scope = NewScope("star:admin:*", "Grants full access to administrator-related permissions")

	// Admin user scopes, must be granted
	ScopeAdminReadUsers Scope = NewScope("star:admin:read:users", "Permits reading all user-related information")
	ScopeAdminModifyUsers Scope = NewScope("star:admin:update:users", "Permits updating any user details")
	ScopeAdminDeleteUsers Scope = NewScope("star:admin:delete:users", "Permits deleting any user's accounts")

	// Admin role scopes, must be granted
	ScopeAdminReadRoles Scope = NewScope("star:admin:read:roles", "Permits reading all roles")
	ScopeAdminModifyRoles Scope = NewScope("star:admin:update:roles", "Permits modifying existing roles")
	ScopeAdminCreateRoles Scope = NewScope("star:admin:create:roles", "Permits creating new roles")
	ScopeAdminDeleteRoles Scope = NewScope("star:admin:delete:roles", "Permits deleting roles")

	// Admin app scopes, must be granted
	ScopeAdminReadApps Scope = NewScope("star:admin:read:apps", "Permits viewing all application information")
	ScopeAdminModifyApps Scope = NewScope("star:admin:update:apps", "Permits modifying applications")
	ScopeAdminCreateApps Scope = NewScope("star:admin:create:apps", "Permits creating new applications")
	ScopeAdminDeleteApps Scope = NewScope("star:admin:delete:apps", "Permits deleting applications")
)

// The list of all built-in default scopes
var DefaultScopes = []Scope{
	ScopeAll, ScopeUserAll, ScopeReadEmail, ScopeReadSessions, ScopeDeleteSessions, ScopeDeleteUser, ScopeChangePassword,
	ScopeReadMetrics,
	ScopeAdminAll,
	ScopeAdminReadUsers, ScopeAdminModifyUsers, ScopeAdminDeleteUsers,
	ScopeAdminReadRoles, ScopeAdminModifyRoles, ScopeAdminCreateRoles, ScopeAdminDeleteRoles,
	ScopeAdminReadApps, ScopeAdminModifyApps, ScopeAdminCreateApps, ScopeAdminDeleteApps,
}

// Check this the scope permits another, 'other' should always be the comparison while 's' is the
// scope the user owns to support wildcards
func Permits(permission string, other string) bool {
	if permission == other {
		return true
	}

	if strings.HasSuffix(string(permission), "*") {
		prefix := strings.TrimSuffix(string(permission), "*")
		if strings.HasPrefix(string(other), prefix) {
			return true
		}
	}

	return false
}