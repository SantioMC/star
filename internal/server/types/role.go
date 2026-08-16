package types

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=64"`
	Description string   `json:"description" validate:"omitempty,max=255"`
	Permissions []string `json:"permissions" validate:"omitempty"`
}

type UpdateRoleRequest struct {
	Name        *string   `json:"name" validate:"omitempty,min=1,max=64"`
	Description *string   `json:"description" validate:"omitempty,max=255"`
	Permissions *[]string `json:"permissions" validate:"omitempty"`
}
