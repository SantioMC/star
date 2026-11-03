package server

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/passkey"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
)

func mfaRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFARead), getMfaFactors)
	router.Delete("/totp", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionDeleteMFA), deleteTotp)
	router.Post("/totp/enroll", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), enrollTotp)
	router.Post("/totp/verify", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), verifyTotp)
	router.Get("/backup", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFARead), getBackupCodes)
	router.Post("/backup", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), regenerateBackupCodes)
	router.Get("/passkeys", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFARead), getPasskeys)
	router.Get("/passkeys/signal", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFARead), getPasskeySignal)
	router.Post("/passkeys/enroll", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), enrollPasskey)
	router.Post("/passkeys/verify", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), verifyPasskey)
	router.Patch("/passkeys/:passkey", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionEnrollMFA), renamePasskey)
	router.Delete("/passkeys/:passkey", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserMFAWrite), middleware.RequirePermission(permission.PermissionDeleteMFA), deletePasskey)
}

func getMfaFactors(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	factors, err := mfa.GetMFA(ctx, star, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"factors": factors,
	})
}

func deleteTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, userID)
	if totp == nil {
		return fiber.NewError(400, "TOTP is not currently enrolled")
	}

	if err := star.Database.DeleteTotpByID(ctx, totp.ID); err != nil {
		return err
	}

	if err := mfa.RemoveOrphanedBackupCodes(ctx, star, userID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func enrollTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, user.ID)
	if totp != nil {
		if totp.Verified {
			return fiber.NewError(400, "TOTP is already enrolled")
		}

		// Regenerate a code for the user
		if err := star.Database.DeleteTotpByID(ctx, totp.ID); err != nil {
			return err
		}
	}

	secret, uri, err := mfa.CreateTotp(star, user)
	if err != nil {
		return err
	}

	totp = &database.Totp{
		ID:        database.CreateID(database.TotpKey),
		UserID:    user.ID,
		Secret:    secret,
		Verified:  false,
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateTotp(ctx, totp); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"uri":    uri,
		"secret": secret,
	})
}

func verifyTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, userID)
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

	if err := mfa.VerifyTotp(ctx, star, totp, body.Code); errors.Is(err, mfa.ErrInvalidTotp) {
		return fiber.NewError(400, "Invalid authentication code, please try again")
	} else if err != nil {
		return err
	}

	totp.Verified = true
	if err := star.Database.UpdateTotp(ctx, totp); err != nil {
		return err
	}

	codes, err := mfa.IssueInitialBackupCodes(ctx, star, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success":      true,
		"backup_codes": codes,
	})
}

func getBackupCodes(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	remaining, err := star.Database.CountBackupCodes(ctx, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"remaining": remaining,
	})
}

func regenerateBackupCodes(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	enabled, err := mfa.HasSecondFactor(ctx, star, userID)
	if err != nil {
		return err
	} else if !enabled {
		return fiber.NewError(400, "An authenticator app or passkey must be enabled before creating backup codes")
	}

	codes, err := mfa.RegenerateBackupCodes(ctx, star, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"backup_codes": codes,
	})
}

func getPasskeys(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	passkeys, err := star.Database.GetUserPasskeys(ctx, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformPasskeys(passkeys))
}

func getPasskeySignal(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	passkeys, err := star.Database.GetUserPasskeys(ctx, user.ID)
	if err != nil {
		return err
	}

	credentialIDs := make([]string, 0, len(passkeys))
	for _, value := range passkeys {
		credentialIDs = append(credentialIDs, base64.RawURLEncoding.EncodeToString(value.CredentialID))
	}

	return ctx.JSON(fiber.Map{
		"rp_id":          star.WebAuthn.Config.RPID,
		"user_id":        base64.RawURLEncoding.EncodeToString(passkey.UserHandle(user)),
		"name":           user.Name,
		"display_name":   user.Name,
		"credential_ids": credentialIDs,
	})
}

func enrollPasskey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	challenge, options, err := passkey.BeginRegistration(ctx, star, user)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"challenge": challenge,
		"options":   options,
	})
}

func verifyPasskey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	var body types.VerifyPasskeyRequest
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		return fiber.NewError(400, "A name is required for your passkey")
	}

	created, err := passkey.FinishRegistration(ctx, star, user, body.Challenge, name, body.Credential)
	if err != nil {
		return err
	}

	codes, err := mfa.IssueInitialBackupCodes(ctx, star, user.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"passkey":      transformer.TransformPasskey(*created),
		"backup_codes": codes,
	})
}

func renamePasskey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	var body types.RenamePasskeyRequest
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		return fiber.NewError(400, "A name is required for your passkey")
	}

	existing := star.Database.GetUserPasskeyByID(ctx, userID, ctx.Params("passkey"))
	if existing == nil {
		return fiber.NewError(404, "Passkey not found")
	}

	existing.Name = name
	if err := star.Database.RenamePasskey(ctx, existing); err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformPasskey(*existing))
}

func deletePasskey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	existing := star.Database.GetUserPasskeyByID(ctx, userID, ctx.Params("passkey"))
	if existing == nil {
		return fiber.NewError(404, "Passkey not found")
	}

	if err := star.Database.DeletePasskeyByID(ctx, existing.ID); err != nil {
		return err
	}

	if err := mfa.RemoveOrphanedBackupCodes(ctx, star, userID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
