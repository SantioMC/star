package email

import (
	"html"
	"strings"

	"github.com/tubblelive/star/internal/cryptography"
)

type EmailProvider interface {
	GetSender() string
	Send(to string, subject string, body string, options map[string]string) error
}

func parse(body string, options map[string]string) string {
	for key, value := range options {
		body = strings.ReplaceAll(body, "{{"+key+"}}", html.EscapeString(value))
	}

	return body
}

func CreateCode(length int) string {
	return cryptography.CreateSecureStringFrom("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", length)
}
