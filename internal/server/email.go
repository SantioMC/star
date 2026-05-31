package server

import (
	"crypto/subtle"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
)

func emailRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionReadEmails), getUserEmails)
	router.Post("/", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAddEmails), addUserEmail)

	router.Post("/verify", limiter.New(middleware.HeavyRateLimiter), verifyEmail)

	router.Delete("/:email", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionDeleteEmails), deleteUserEmail)
	router.Post("/:email/primary", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionUpdatePrimaryEmail), updatePrimaryEmail)
}

func getUserEmails(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	emails := star.Database.GetEmailsByUserID(ctx, user.ID)
	return ctx.JSON(transformer.TransformEmails(emails))
}

func addUserEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.AddUserEmailRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	// Prevent taking over an email that is either verified, or has recently been added within the last hour
	existing := star.Database.GetEmailByName(ctx, data.Email)
	canUse := existing == nil || !existing.Verified || existing.CreatedAt.Add(time.Hour).Before(time.Now())

	if !canUse {
		return fiber.NewError(400, "the provided email is either already in use or has recently been added to an account")
	}

	email := &database.UserEmail{
		ID:        database.CreateID(database.EmailKey),
		UserID:    user.ID,
		Email:     data.Email,
		Verified:  false,
		Primary:   false,
		Code:      email.CreateCode(star.Settings.EmailCodeLength),
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateEmail(ctx, email); err != nil {
		return err
	}

	// Send verification email
	url := star.Config.ServerURL + "/verify?email=" + data.Email + "&code=" + email.Code
	if err := star.Email.Send(
		data.Email,
		"Verify your email address",
		emails.VerificationEmail,
		map[string]string{
			"Name":            user.Name,
			"VerificationURL": url,
			"Code":            email.Code,
		},
	); err != nil {
		slog.Warn("failed to send verification email", "err", err, "email", data.Email, "code", email.Code)
	}

	return ctx.JSON(transformer.TransformEmail(email))
}

func deleteUserEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != user.ID {
		return fiber.NewError(401, "you are not permitted to delete this email")
	}

	emails := star.Database.GetEmailsByUserID(ctx, user.ID)
	if len(emails) == 1 {
		return fiber.NewError(400, "you may not remove the only email on your account")
	}

	if err := star.Database.DeleteEmail(ctx, emailID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func updatePrimaryEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != user.ID {
		return fiber.NewError(401, "the provided email is not associated with your account")
	}

	if err := star.Database.SetPrimaryEmail(ctx, user.ID, emailID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func verifyEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.VerifyEmailRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	email := star.Database.GetEmailByName(ctx, request.Email)
	if email == nil || email.Verified {
		return fiber.NewError(404, "email not found")
	}

	if subtle.ConstantTimeCompare([]byte(email.Code), []byte(request.Code)) != 1 {
		return fiber.NewError(400, "invalid code specified")
	}

	email.Verified = true
	if err := star.Database.UpdateEmail(ctx, email); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
