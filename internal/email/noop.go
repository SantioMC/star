package email

import (
	"html/template"
	"log/slog"
)

type NoopProvider struct{}

func (provider *NoopProvider) Send(to string, subject string, body *template.Template, data any) error {
	slog.Info("email not setup, printing details", "subject", subject, "to", to, "data", data)
	return nil
}

func (provider *NoopProvider) GetSender() string {
	return ""
}
