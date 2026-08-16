package types

type VerifyTotpRequest struct {
	Code string `json:"code" validate:"required,min=6,max=32"`
}

type LoginFlowTotpRequest struct {
	Code string `json:"code" validate:"required,min=6,max=32"`
	Flow string `json:"flow" validate:"required,max=64"`
}
