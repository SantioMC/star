package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
)

func mfaRouter(router fiber.Router) {
	router.Delete("/totp", middleware.RequirePermission(permission.PermissionDeleteMFA), deleteTotp)
	router.Post("/totp/enroll", middleware.RequirePermission(permission.PermissionEnrollMFA), enrollTotp)
	router.Post("/totp/verify", middleware.RequirePermission(permission.PermissionEnrollMFA), verifyTotp)
}

func deleteTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	mfa := star.Database.GetUserTotp(ctx, userId)
	if mfa == nil {
		return fiber.NewError(400, "TOTP is not currently enrolled")
	}

	if err := star.Database.DeleteTotpByID(ctx, mfa.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func enrollTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, userId)
	if totp != nil {
		if totp.Verified {
			return fiber.NewError(400, "TOTP is already enrolled")
		}

		// Regenerate a code for the user
		if err := star.Database.DeleteTotpByID(ctx, totp.ID); err != nil {
			return err
		}
	}

	secret, err := mfa.CreateTotp(star, userId)
	if err != nil {
		return err
	}

	totp = &database.Totp{
		ID:        database.CreateID(database.TotpKey),
		UserID:    userId,
		Secret:    secret,
		Verified:  false,
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateTotp(ctx, totp); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"secret": secret,
	})
}

func verifyTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, userId)
	if totp == nil {
		return fiber.NewError(400, "totp not enrolled")
	}

	if totp.Verified {
		// todo: reuse flow as completion?
		return fiber.NewError(400, "totp is already verified")
	}

	var body types.VerifyTotpRequest
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	if err := mfa.VerifyTotp(totp.Secret, body.Code); err != nil {
		return err
	}

	totp.Verified = true
	if err := star.Database.UpdateTotp(ctx, totp); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
