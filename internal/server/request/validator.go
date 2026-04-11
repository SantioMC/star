package request

import (
	"github.com/go-playground/validator/v10"
)

type StructValidator struct{}

var validate *validator.Validate = validator.New(
	validator.WithRequiredStructEnabled(),
)

func NewStructValidator() *StructValidator {
	return &StructValidator{}
}

func (v *StructValidator) Validate(out any) error {
	return validate.Struct(out)
}
