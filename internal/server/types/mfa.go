package types

import "encoding/json"

type VerifyTotpRequest struct {
	Code string `json:"code" validate:"required,min=6,max=32"`
}

type LoginFlowCodeRequest struct {
	Code string `json:"code" validate:"required,min=6,max=32"`
	Flow string `json:"flow" validate:"required,max=64"`
}

type VerifyPasskeyRequest struct {
	Challenge  string          `json:"challenge" validate:"required,max=64"`
	Name       string          `json:"name" validate:"required,max=64"`
	Credential json.RawMessage `json:"credential" validate:"required"`
}

type RenamePasskeyRequest struct {
	Name string `json:"name" validate:"required,max=64"`
}

type PasskeyOptionsRequest struct {
	Flow string `json:"flow" validate:"omitempty,max=64"`
}

type PasskeyLoginRequest struct {
	Challenge  string          `json:"challenge" validate:"required,max=64"`
	Flow       string          `json:"flow" validate:"omitempty,max=64"`
	Credential json.RawMessage `json:"credential" validate:"required"`
}
