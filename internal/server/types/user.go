package types

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username  string  `json:"username" validate:"required"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

type ChangePasswordRequest struct {
	OldPassword *string `json:"old_password" validate:"omitempty"`
	NewPassword string  `json:"new_password" validate:"required"`
}

type AddUserEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}
