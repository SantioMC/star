package email

import (
	"math/rand"
	"strings"
)

type EmailProvider interface {
	GetSender() string
	Send(to string, subject string, body string, options map[string]string) error
}

func parse(html string, options map[string]string) string {
	for key, value := range options {
		html = strings.ReplaceAll(html, "{{"+key+"}}", value)
	}

	return html
}

var alphabet = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// This is ONLY for email verification, not to be used for anything else
func CreateCode(length int) string {
	b := make([]rune, length)
	repeats := 0

	for i := range b {
		if i != 0 && repeats < 3 && rand.Intn(10) <= 2 {
			b[i] = b[i-1]
			repeats++
		} else {
			b[i] = alphabet[rand.Intn(len(alphabet))]
			repeats = 0
		}
	}

	return string(b)
}
