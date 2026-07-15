package types

type VerifyTotpRequest struct {
	Code string `json:"code" validate:"required,min=6,max=32"`
}

type LoginFlowTotpRequest struct {
	Code   string `json:"code" validate:"required,min=6,max=32"`
	FlowID string `json:"flow" validate:"required,max=32"`
}
