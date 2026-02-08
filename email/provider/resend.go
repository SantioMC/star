package provider

import (
	"github.com/resend/resend-go/v3"
)

type ResendProvider struct {
	sender string
	client *resend.Client
}

func NewResendProvider(sender string, apiKey string) *ResendProvider {
	client := resend.NewClient(apiKey)

	return &ResendProvider{
		client: client,
	}
}

func (provider *ResendProvider) Send(to string, subject string, body string) error {
	params := &resend.SendEmailRequest{
		From:    provider.sender,
		To:      []string{to},
		Html:    body,
		Subject: subject,
	}

	_, err := provider.client.Emails.Send(params)
	return err
}

func (provider *ResendProvider) GetSender() string {
	return provider.sender
}
