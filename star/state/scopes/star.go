package scopes

import "strings"

type Scope string

const (
	ScopeAll Scope = "star:*"
	
	// User scopes, given to all users (starts with `star:user:`)
	ScopeUserAll Scope = "star:user:*"
	ScopeReadEmail Scope = "star:user:email"
	ScopeReadSessions Scope = "star:user:read:sessions"
	ScopeDeleteSessions Scope = "star:user:delete:sessions"
	ScopeDeleteUser Scope = "star:user:delete"
	ScopeChangePassword Scope = "star:user:change_password"

	// Additional scopes, must be granted
	ScopeReadMetrics Scope = "star:server:metrics"

	// Generic admin scopes, must be granted
	ScopeAdminAll Scope = "star:admin:*"

	// Admin user scopes, must be granted
	ScopeAdminReadUsers Scope = "star:admin:read:users"
	ScopeAdminModifyUsers Scope = "star:admin:update:users"
	ScopeAdminDeleteUsers Scope = "star:admin:delete:users"

	// Admin role scopes, must be granted
	ScopeAdminReadRoles Scope = "star:admin:read:roles"

	// Admin app scopes, must be granted
	ScopeAdminReadApps Scope = "star:admin:read:apps"
)

// The list of all built-in default scopes
var DefaultScopes = []Scope{
	ScopeAll, ScopeUserAll, ScopeReadEmail, ScopeDeleteUser,
}

// Check this the scope permits another, 'other' should always be the comparison while 's' is the
// scope the user owns to support wildcards
func (s Scope) Permits(other Scope) bool {
	if s == other {
		return true
	}

	if strings.HasSuffix(string(s), "*") {
		prefix := strings.TrimSuffix(string(s), "*")
		if strings.HasPrefix(string(other), prefix) {
			return true
		}
	}

	return false
}