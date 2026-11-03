package server

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/types"
)

func createEmailPreview(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.EmailPreviewRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	current, err := currentBranding(star, ctx)
	if err != nil {
		return err
	}

	if err := applyBranding(current, &request.Branding); err != nil {
		return err
	}

	payload, err := json.Marshal(types.EmailPreviewTicket{
		EmailPreviewRequest: request,
		ExpiresAt:           time.Now().Add(time.Minute * 5).Unix(),
	})

	if err != nil {
		return err
	}

	return ctx.JSON(types.EmailPreviewResponse{
		Ticket: cryptography.Sign("email-preview", payload, star.Config.AppToken),
	})
}

func renderEmailPreview(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	payload, valid := cryptography.Verify("email-preview", ctx.Query("ticket"), star.Config.AppToken)
	if !valid {
		return fiber.NewError(400, "Invalid email preview ticket")
	}

	var ticket types.EmailPreviewTicket
	if err := json.Unmarshal(payload, &ticket); err != nil {
		return err
	}

	if time.Now().Unix() > ticket.ExpiresAt {
		return fiber.NewError(410, "This email preview has expired")
	}

	current, err := currentBranding(star, ctx)
	if err != nil {
		return err
	}

	if err := applyBranding(current, &ticket.Branding); err != nil {
		return err
	}

	if ticket.RemoveLogo {
		current.LogoUpdatedAt = nil
	}

	brand := branding.ForEmail(current, star.Config.ServerURL)
	body, data, found := emails.Sample(ticket.Template, brand, star.Config.ServerURL)

	if !found {
		return fiber.NewError(404, "Unknown email template")
	}

	html, err := email.Render(body, data)
	if err != nil {
		return err
	}

	ancestors := append([]string{"'self'"}, star.Config.Origins...)
	csp := []string{
		"default-src 'none'",
		"style-src 'unsafe-inline' https://fonts.googleapis.com",
		"font-src https://fonts.gstatic.com",
		"img-src 'self' data:",
		"frame-ancestors " + strings.Join(ancestors, " "),
		"sandbox",
	}

	ctx.Set(fiber.HeaderContentSecurityPolicy, strings.Join(csp, "; "))
	ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	ctx.Set(fiber.HeaderReferrerPolicy, "no-referrer")
	ctx.Set(fiber.HeaderCacheControl, "no-store")
	ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)

	return ctx.SendString(html)
}
