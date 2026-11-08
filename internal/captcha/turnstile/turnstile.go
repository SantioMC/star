package turnstile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/outbound"
)

var (
	ErrInvalidSecret   = fiber.NewError(400, "Cloudflare rejected the secret key")
	ErrInvalidResponse = fiber.NewError(400, "Cloudflare rejected the captcha response, make sure the site key and secret key belong to the same widget")
	ErrExpiredResponse = fiber.NewError(400, "The captcha response expired or was already used, solve it again")
	ErrUnavailable     = fiber.NewError(502, "Cloudflare Turnstile could not be reached")
)

type TurnstileValidator struct {
	siteKey string
	secret  string
}

type Request struct {
	Secret string `json:"secret"`
	Token  string `json:"response"`
}

type Response struct {
	Success    bool     `json:"success"`
	Timestamp  string   `json:"challenge_ts"`
	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
	Action     string   `json:"action"`
}

func NewTurnstileValidator(siteKey string, secret string) *TurnstileValidator {
	return &TurnstileValidator{
		siteKey: siteKey,
		secret:  secret,
	}
}

func (v *TurnstileValidator) GetSiteKey() string {
	return v.siteKey
}

func (v *TurnstileValidator) ValidateToken(token string) bool {
	return v.Check(token) == nil
}

func (v *TurnstileValidator) Check(token string) error {
	if strings.TrimSpace(token) == "" || len(token) > 2048 {
		return ErrInvalidResponse
	}

	data, err := v.verify(token)
	if err != nil {
		slog.Error("failed to verify captcha token", "err", err)
		return ErrUnavailable
	}

	if data.Success {
		return nil
	}

	switch {
	case slices.Contains(data.ErrorCodes, "missing-input-secret"), slices.Contains(data.ErrorCodes, "invalid-input-secret"):
		return ErrInvalidSecret
	case slices.Contains(data.ErrorCodes, "timeout-or-duplicate"):
		return ErrExpiredResponse
	default:
		return ErrInvalidResponse
	}
}

func (v *TurnstileValidator) verify(token string) (*Response, error) {
	body, err := json.Marshal(Request{
		Secret: v.secret,
		Token:  token,
	})

	if err != nil {
		return nil, err
	}

	response, err := outbound.Public.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	data := Response{}
	if err := json.Unmarshal(contents, &data); err != nil {
		return nil, fmt.Errorf("siteverify returned status %d: %s", response.StatusCode, contents)
	}

	return &data, nil
}
