package server

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func emailRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserEmailRead), middleware.RequirePermission(permission.PermissionReadEmails), getUserEmails)
	router.Post("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserEmailWrite), middleware.RequirePermission(permission.PermissionAddEmails), addUserEmail)

	router.Post("/verify", middleware.RateLimit(10), middleware.OptionalAuth, verifyEmail)
	router.Post("/resend", middleware.RateLimit(10), resendVerification)

	router.Delete("/:email", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserEmailWrite), middleware.RequirePermission(permission.PermissionDeleteEmails), deleteUserEmail)
	router.Post("/:email/primary", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserEmailWrite), middleware.RequirePermission(permission.PermissionUpdatePrimaryEmail), updatePrimaryEmail)
}

func getUserEmails(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	userId, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	emails := star.Database.GetEmailsByUserID(ctx, userId)
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

	settings := star.Settings.Load()
	if err := settings.GetEmailDomainPolicy().Check(data.Email); err != nil {
		return err
	}

	claimedError := fiber.NewError(400, "the provided email is either already in use or has recently been added to an account")
	if star.Database.EmailClaimed(ctx, data.Email) {
		return claimedError
	}

	userEmail := newUserEmail(settings, user.ID, data.Email, false, time.Now())

	err = star.Database.CreateEmail(ctx, userEmail)
	if errors.Is(err, database.ErrDuplicate) {
		return claimedError
	} else if err != nil {
		return err
	}

	if !userEmail.Verified {
		sendVerificationEmail(star, user.Name, userEmail.Email, userEmail.Code)
	}

	audit.Log(ctx, audit.EmailAdd, audit.With("email_id", userEmail.ID), audit.Private("email", userEmail.Email))

	return ctx.JSON(transformer.TransformEmail(userEmail))
}

func newUserEmail(settings *database.Settings, userID string, address string, primary bool, created time.Time) *database.UserEmail {
	return &database.UserEmail{
		ID:           database.CreateID(database.EmailKey),
		UserID:       userID,
		Email:        address,
		Verified:     !settings.EmailVerification,
		SelfAsserted: !settings.EmailVerification,
		Primary:      primary,
		Code:         email.CreateCode(settings.EmailCodeLength),
		CreatedAt:    created,
	}
}

func sendVerificationEmail(star *state.StarServer, name string, address string, code string) {
	verificationURL := emails.VerificationURL(star.Config.ServerURL, address, code)
	err := star.Email.Send(
		address,
		"Verify your email address",
		emails.Verification,
		emails.VerificationData{
			Brand:           branding.ForEmail(star.Branding.Load(), star.Config.ServerURL),
			Name:            name,
			VerificationURL: verificationURL,
			Code:            code,
		},
	)

	if err != nil {
		slog.Warn("failed to send verification email", "err", err, "email", address, "code", code)
	}
}

func resendVerification(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.ResendVerificationRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	success := fiber.Map{
		"success": true,
	}

	userEmail := star.Database.GetEmailByName(ctx, request.Email)
	if userEmail == nil || userEmail.Verified {
		return ctx.JSON(success)
	}

	user := star.Database.GetUserByID(ctx, userEmail.UserID)
	if user == nil {
		return ctx.JSON(success)
	}

	code := email.CreateCode(star.Settings.Load().EmailCodeLength)
	reset, err := star.Database.ResetEmailCode(ctx, userEmail.ID, code)

	if err != nil {
		return err
	} else if reset {
		sendVerificationEmail(star, user.Name, userEmail.Email, code)
	}

	return ctx.JSON(success)
}

func deleteUserEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	userId, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != userId {
		return fiber.NewError(401, "you are not permitted to delete this email")
	}

	emails := star.Database.GetEmailsByUserID(ctx, userId)
	if len(emails) == 1 {
		return fiber.NewError(400, "you may not remove the only email on your account")
	}

	if err := star.Database.DeleteEmail(ctx, emailID); err != nil {
		return err
	}

	audit.Log(ctx, audit.EmailRemove, audit.With("email_id", email.ID), audit.Private("email", email.Email))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func updatePrimaryEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	userId, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != userId {
		return fiber.NewError(401, "the provided email is not associated with your account")
	}

	if err := star.Database.SetPrimaryEmail(ctx, userId, emailID); err != nil {
		return err
	}

	audit.Log(ctx, audit.EmailPrimary, audit.With("email_id", email.ID), audit.Private("email", email.Email))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func verifyEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.VerifyEmailRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	var requesterID *string
	if jwt, err := request.GetJWT(ctx); err == nil {
		requesterID = &jwt.Subject
	}

	verified, err := star.Database.VerifyEmailCode(ctx, data.Email, data.Code, requesterID)
	if errors.Is(err, database.ErrEmailOwnerRequired) {
		return fiber.NewError(403, "Sign in to the account this email was added to, then open the verification link again")
	} else if err != nil {
		return err
	} else if !verified {
		return fiber.NewError(400, "This verification code is invalid or has expired, please request a new one")
	}

	if verifiedEmail := star.Database.GetEmailByName(ctx, data.Email); verifiedEmail != nil {
		audit.Log(ctx, audit.EmailVerify, audit.ByID(verifiedEmail.UserID), audit.With("email_id", verifiedEmail.ID), audit.Private("email", verifiedEmail.Email))
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
