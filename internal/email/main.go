package email

import (
	"bytes"
	"html/template"

	"github.com/tubblelive/star/internal/cryptography"
)

type EmailProvider interface {
	GetSender() string
	Send(to string, subject string, body *template.Template, data any) error
}

func Render(body *template.Template, data any) (string, error) {
	var buffer bytes.Buffer
	if err := body.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func CreateCode(length int) string {
	return cryptography.CreateSecureStringFrom("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", length)
}
