package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type GrantedScope struct {
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

type AuthorizedApplication struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Scopes       []GrantedScope `json:"scopes"`
	AuthorizedAt time.Time      `json:"authorized_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastUsedAt   *time.Time     `json:"last_used_at"`
}

func TransformGrant(grant database.UserApplicationGrant, descriptions map[string]string) AuthorizedApplication {
	scopes := make([]GrantedScope, 0, len(grant.Scopes))
	for _, scope := range grant.Scopes {
		description, found := descriptions[scope]
		if !found {
			continue
		}

		scopes = append(scopes, GrantedScope{
			Scope:       scope,
			Description: description,
		})
	}

	return AuthorizedApplication{
		ID:           grant.ApplicationID,
		Name:         grant.ApplicationName,
		URL:          grant.ApplicationURL,
		Scopes:       scopes,
		AuthorizedAt: grant.CreatedAt,
		UpdatedAt:    grant.UpdatedAt,
		LastUsedAt:   grant.LastUsedAt,
	}
}

func TransformGrants(grants []database.UserApplicationGrant, descriptions map[string]string) []AuthorizedApplication {
	transformed := make([]AuthorizedApplication, len(grants))
	for idx, grant := range grants {
		transformed[idx] = TransformGrant(grant, descriptions)
	}

	return transformed
}
