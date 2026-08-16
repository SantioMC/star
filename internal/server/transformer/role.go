package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type AdminRole struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Internal    bool      `json:"internal"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

func TransformRole(role database.Role, permissions []database.Permission) AdminRole {
	names := make([]string, len(permissions))
	for idx, permission := range permissions {
		names[idx] = permission.Name
	}

	return AdminRole{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Internal:    role.Internal,
		Permissions: names,
		CreatedAt:   role.CreatedAt,
	}
}
