package server

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

var ErrPasswordResetInvalid = fiber.NewError(410, "This password reset link is invalid or has expired, please request a new one")

func passwordResetRouter(router fiber.Router) {
	router.Post("/", middleware.RateLimit(5), middleware.RequireCaptcha, requestPasswordReset)
	router.Post("/verify", middleware.RateLimit(10), verifyPasswordReset)
	router.Post("/complete", middleware.RateLimit(10), completePasswordReset)
}

func requestPasswordReset(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.RequestPasswordResetRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	success := fiber.Map{
		"success": true,
	}

	userEmail := star.Database.GetEmailByName(ctx, data.Email)
	if userEmail == nil || !userEmail.Verified {
		return ctx.JSON(success)
	}

	if !star.Database.UserHasPermission(ctx, userEmail.UserID, permission.PermissionResetPassword.Name) {
		slog.Info("password reset denied", "reason", "missing permission", "user", userEmail.UserID)
		return ctx.JSON(success)
	}

	account := star.Database.GetInternalAccountByUserID(ctx, userEmail.UserID)
	if account == nil || account.Password == nil {
		return ctx.JSON(success)
	}

	user := star.Database.GetUserByID(ctx, userEmail.UserID)
	if user == nil {
		return ctx.JSON(success)
	}

	created := time.Now()
	reset := &database.PasswordReset{
		ID:        database.CreateID(database.PasswordResetKey),
		UserID:    user.ID,
		Token:     database.CreatePasswordResetToken(),
		ExpiresAt: created.Add(time.Minute * 30),
		CreatedAt: created,
	}

	issued, err := star.Database.IssuePasswordReset(ctx, reset)
	if err != nil {
		return err
	} else if !issued {
		return ctx.JSON(success)
	}

	slog.Info("password reset requested", "user", user.ID)
	go sendPasswordResetEmail(star, user.Name, userEmail.Email, reset.Token, star.ClientIP(ctx))

	return ctx.JSON(success)
}

func sendPasswordResetEmail(star *state.StarServer, name string, address string, token string, ipAddress string) {
	err := star.Email.Send(
		address,
		"Reset your password",
		emails.PasswordReset,
		emails.PasswordResetData{
			Brand:     branding.ForEmail(star.Branding.Load(), star.Config.ServerURL),
			Name:      name,
			ResetURL:  emails.PasswordResetURL(star.Config.ServerURL, token),
			IPAddress: ipAddress,
		},
	)

	if err != nil {
		slog.Warn("failed to send password reset email", "err", err, "email", address)
	}
}

func notifyPasswordChanged(ctx fiber.Ctx, star *state.StarServer, userID string, ipAddress string) {
	user := star.Database.GetUserByID(ctx, userID)
	if user == nil {
		return
	}

	primary := star.Database.GetPrimaryEmail(ctx, userID)
	if primary == nil || !primary.Verified {
		return
	}

	data := emails.PasswordChangedData{
		Brand:       branding.ForEmail(star.Branding.Load(), star.Config.ServerURL),
		Name:        user.Name,
		ChangedAt:   emails.FormatTime(time.Now()),
		IPAddress:   ipAddress,
		RecoveryURL: emails.RecoveryURL(star.Config.ServerURL),
	}

	go func() {
		if err := star.Email.Send(primary.Email, "Your password was changed", emails.PasswordChanged, data); err != nil {
			slog.Warn("failed to send password changed email", "err", err, "email", primary.Email)
		}
	}()
}

func pendingPasswordReset(ctx fiber.Ctx, star *state.StarServer, token string) (*database.PasswordReset, error) {
	reset := star.Database.GetPasswordResetByToken(ctx, token)
	if reset == nil {
		return nil, ErrPasswordResetInvalid
	}

	if !star.Database.UserHasPermission(ctx, reset.UserID, permission.PermissionResetPassword.Name) {
		return nil, fiber.NewError(403, "Password resets are not permitted for this account")
	}

	return reset, nil
}

func verifyPasswordReset(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.VerifyPasswordResetRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	if _, err := pendingPasswordReset(ctx, star, data.Token); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func completePasswordReset(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.CompletePasswordResetRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	reset, err := pendingPasswordReset(ctx, star, data.Token)
	if err != nil {
		return err
	}

	sanitized, err := password.ValidatePassword(data.Password, star.Settings.Load().GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(sanitized)

	completed, err := star.Database.CompletePasswordReset(ctx, reset, hashed, alg)
	if err != nil {
		return err
	} else if !completed {
		return ErrPasswordResetInvalid
	}

	cache.UpdateRevocations(ctx, star)

	ipAddress := star.ClientIP(ctx)
	slog.Info("password reset completed", "user", reset.UserID, "ip_address", ipAddress)
	notifyPasswordChanged(ctx, star, reset.UserID, ipAddress)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
