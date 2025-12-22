package scopes

import "strings"

type Scope string

const (
	ScopeAll Scope = "star:*"
	ScopeUserAll Scope = "star:user:*"
	ScopeReadEmail Scope = "star:user:email"
	ScopeDeleteUser Scope = "star:user:delete"
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

	parts := strings.Split(string(s), ":")
	otherParts := strings.Split(string(other), ":")

	for idx, part := range parts {
		otherPart := otherParts[idx]
		
		if part == "*" {
			return true
		}

		if part != otherPart {
			return false
		}
	}

	// Must make sure we're in the same level, foo:bar != foo:bar:baz
	return len(parts) == len(otherParts)
}