package email

type EmailProvider interface {
	GetSender() string
	Send(to string, subject string, body string) error
}
