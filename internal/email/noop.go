package email

import "log/slog"

type NoopProvider struct{}

func (provider *NoopProvider) Send(to string, subject string, body string, options map[string]string) error {
	slog.Info("email not setup, printing details", "subject", subject, "to", to, "data", options)
	return nil
}

func (provider *NoopProvider) GetSender() string {
	return ""
}
