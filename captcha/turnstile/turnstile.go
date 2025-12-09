package turnstile

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type TurnstileValidator struct {
	secret string
}

type Request struct {
	Secret string `json:"secret"`
	Token string `json:"response"`
}

type Response struct {
	Success bool `json:"success"`
	Timestamp string `json:"challenge_ts"`
	Hostname string `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
	Action string `json:"action"`
}

func NewTurnstileValidator(secret string) TurnstileValidator {
	return TurnstileValidator{
		secret: secret,
	}
}

func (v *TurnstileValidator) ValidateToken(token string) bool {
	if (len(token) > 2048) {
		slog.Warn("captcha token provided was too large")
		return false
	}

	body, err := json.Marshal(Request{
		Secret: v.secret,
		Token: token,
	})

	if err != nil {
		slog.Error("failed to create json body for captcha", "err", err)
		return false
	}

	response, err := http.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewBuffer(body),
	)
	
	if err != nil {
		slog.Error("failed to verify captcha token", "err", err)
		return false
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read response body", "err", err)
		return false
	}

	if response.StatusCode != 200 {
		slog.Error("failed to verify captcha token", "token", token, "response", string(bytes))
		return false
	}

	data := Response{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		slog.Error("failed to parse cloudflare turnstile response", "response", string(bytes))
		return false
	}

	return data.Success
}