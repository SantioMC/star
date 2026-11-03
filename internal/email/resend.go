package email

import (
	"html/template"

	"github.com/resend/resend-go/v3"
)

type ResendProvider struct {
	sender string
	client *resend.Client
}

func NewResendProvider(sender string, apiKey string) *ResendProvider {
	client := resend.NewClient(apiKey)

	return &ResendProvider{
		sender: sender,
		client: client,
	}
}

func (provider *ResendProvider) Send(to string, subject string, body *template.Template, data any) error {
	html, err := Render(body, data)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    provider.sender,
		To:      []string{to},
		Html:    html,
		Subject: subject,
	}

	_, err = provider.client.Emails.Send(params)
	return err
}

func (provider *ResendProvider) GetSender() string {
	return provider.sender
}
