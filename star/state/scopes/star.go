package scopes

import "strings"

type Scope string

const (
	ScopeAll Scope = "star:*"
	ScopeUserAll Scope = "star:user:*"
	ScopeReadEmail Scope = "star:user:email"
	ScopeDeleteUser Scope = "star:user:delete"
)

var DefaultScopes = []Scope{
	ScopeAll, ScopeUserAll, ScopeReadEmail, ScopeDeleteUser,
}

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

	return true
}