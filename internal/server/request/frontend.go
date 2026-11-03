package request

import (
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/frontend"
	"github.com/tubblelive/star/internal/cryptography"
)

var contentSecurityPolicy = strings.TrimSpace(`
default-src 'self';
script-src 'self' https://challenges.cloudflare.com/turnstile/v0/api.js 'nonce-{nonce}';
connect-src 'self';
frame-src https://challenges.cloudflare.com;
frame-ancestors 'none';
form-action 'self';
object-src 'none';
base-uri 'self';
img-src 'self' data: https://*.gravatar.com;
style-src 'self' 'unsafe-hashes' 'sha256-tcbDxjMo+xKqM21aCGYbs/QAJqB7yUXC06oPWDapBgc=' 'sha256-S8qMpvofolR8Mpjy4kQvEm7m1q8clzU4dfDH0AmvZjo=' https://cdnjs.cloudflare.com/ajax/libs/tabler-icons/3.35.0/tabler-icons.min.css https://fonts.googleapis.com/css2;
font-src 'self' https://cdnjs.cloudflare.com https://fonts.gstatic.com
`)

func renderPage(nonce string) (string, error) {
	file, err := frontend.FrontendFS.Open("build/index.html")
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	renderedScript := fmt.Sprintf("<script nonce='%s'>", nonce)
	rendered := strings.Replace(string(data), "<script>", renderedScript, 1)

	return rendered, nil
}

func ShowFrontend(ctx fiber.Ctx) error {
	nonce := cryptography.CreateSecureString(32)
	rendered, err := renderPage(nonce)
	if err != nil {
		return err
	}

	csp := strings.Replace(contentSecurityPolicy, "{nonce}", nonce, 1)

	ctx.Set(fiber.HeaderContentSecurityPolicy, csp)
	ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	ctx.Set(fiber.HeaderXFrameOptions, "DENY")
	ctx.Set(fiber.HeaderReferrerPolicy, "strict-origin-when-cross-origin")
	ctx.Set(fiber.HeaderPermissionsPolicy, "geolocation=(), camera=(), microphone=()")
	ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)

	return ctx.SendString(rendered)
}
