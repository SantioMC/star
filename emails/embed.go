package emails

import (
	"embed"
	"html/template"
	"time"
)

//go:embed *.html
var files embed.FS

var functions = template.FuncMap{
	"raw": func(value string) template.HTML {
		return template.HTML(value)
	},
}

func parse(content string) *template.Template {
	return template.Must(template.New("layout.html").Funcs(functions).ParseFS(files, "layout.html", content))
}

var Verification = parse("verification.html")

var PasswordReset = parse("password-reset.html")

var PasswordChanged = parse("password-changed.html")

var Invite = parse("invite.html")

type Branding struct {
	Name             string
	AccentColor      string
	AccentForeground string
	HeadingFont      string
	BodyFont         string
	FontURL          string
	LogoURL          string
	SupportEmail     string
	PrivacyURL       string
	TermsURL         string
}

type VerificationData struct {
	Brand           Branding
	Name            string
	VerificationURL string
	Code            string
}

type PasswordResetData struct {
	Brand     Branding
	Name      string
	ResetURL  string
	IPAddress string
}

type PasswordChangedData struct {
	Brand       Branding
	Name        string
	ChangedAt   string
	IPAddress   string
	RecoveryURL string
}

type InviteData struct {
	Brand     Branding
	Name      string
	Inviter   string
	InviteURL string
	ExpiresAt string
}

func FormatTime(at time.Time) string {
	return at.UTC().Format("January 2, 2006 at 3:04 PM MST")
}
