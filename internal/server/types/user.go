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

type AdminSetPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type AdminUpdateUserRequest struct {
	Username  *string   `json:"username" validate:"omitempty"`
	FirstName *string   `json:"first_name" validate:"omitempty,max=255"`
	LastName  *string   `json:"last_name" validate:"omitempty,max=255"`
	Roles     *[]string `json:"roles" validate:"omitempty"`
}

type AddUserEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=1"`
}
