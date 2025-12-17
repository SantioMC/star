package oidc

type Claim string

const (
	// Includes the 'email' and 'email_verified' claims
	ClaimEmail Claim = "email"
	// Includes the 'groups' claim (list of role names)
	ClaimGroups Claim = "groups"
	// Includes the 'roles' claim (list of role data)
	ClaimRoles Claim = "star:roles"
)

// Supported claims by Star
var SupportedClaims []Claim = []Claim{
	ClaimEmail, ClaimGroups, ClaimRoles,
}