package types

type UpdatePasswordSettingsRequest struct {
	MinLength  *int `json:"min_length" validate:"omitempty,min=8,max=128"`
	MaxLength  *int `json:"max_length" validate:"omitempty,min=8,max=512"`
	MinSymbols *int `json:"min_symbols" validate:"omitempty,min=0,max=32"`
	MinUpper   *int `json:"min_upper" validate:"omitempty,min=0,max=32"`
	MinLower   *int `json:"min_lower" validate:"omitempty,min=0,max=32"`
}

type UpdateSettingsRequest struct {
	AllowRegistration *bool                          `json:"allow_registration" validate:"omitempty"`
	AllowLogins       *bool                          `json:"allow_logins" validate:"omitempty"`
	EmailCodeLength   *int                           `json:"email_code_length" validate:"omitempty,min=4,max=12"`
	PasswordSettings  *UpdatePasswordSettingsRequest `json:"password_settings" validate:"omitempty"`
}
