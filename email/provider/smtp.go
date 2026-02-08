package provider

import (
	gomail "gopkg.in/gomail.v2"
)

type SMTPProvider struct {
	sender string
	dialer *gomail.Dialer
}

func NewSMTPProvider(
	sender string,
	host string,
	port int,
	username string,
	password string,
) *SMTPProvider {
	dialer := gomail.NewDialer(
		host,
		port,
		username,
		password,
	)

	return &SMTPProvider{
		sender: sender,
		dialer: dialer,
	}
}

func (provider *SMTPProvider) Send(to string, subject string, body string) error {
	message := gomail.NewMessage()

	message.SetHeader("From", provider.sender)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	err := provider.dialer.DialAndSend(message)

	return err
}

func (provider *SMTPProvider) GetSender() string {
	return provider.sender
}
