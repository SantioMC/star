package types

type CreateRevocationRequest struct {
	Kind string  `json:"kind" validate:"required,oneof=user global"`
	User *string `json:"user" validate:"required_if=Kind user,omitempty,max=64"`
}
