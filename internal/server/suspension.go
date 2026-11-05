package server

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
)

func suspensionReason(value string) (string, error) {
	reason := strings.TrimSpace(value)
	if reason == "" {
		return "", fiber.NewError(400, "A reason must be provided")
	}

	return reason, nil
}

func adminGetUserSuspensions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	_, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	before, err := request.GetCursor(ctx)
	if err != nil {
		return err
	}

	suspensions, err := star.Database.GetUserSuspensions(ctx, user.ID, before, limit+1)
	if err != nil {
		return err
	}

	return ctx.JSON(types.NewPage(suspensions, limit, func(last database.UserSuspension) database.Cursor {
		return database.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}))
}

func adminSuspendUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if user.ID == requesterID {
		return fiber.NewError(403, "You are not permitted to suspend your own user")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	var data types.SuspendUserRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	reason, err := suspensionReason(data.Reason)
	if err != nil {
		return err
	}

	if data.ExpiresAt != nil && !data.ExpiresAt.After(time.Now()) {
		return fiber.NewError(400, "The suspension must end in the future")
	}

	err = star.Database.SuspendUser(ctx, &database.UserSuspension{
		ID:          database.CreateID(database.UserSuspensionKey),
		UserID:      user.ID,
		Reason:      reason,
		SuspendedBy: &requesterID,
		ExpiresAt:   data.ExpiresAt,
		CreatedAt:   time.Now(),
	})

	if errors.Is(err, database.ErrDuplicate) {
		return fiber.NewError(400, "This user is already suspended")
	} else if err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	slog.Info("user suspended", "user", user.ID, "by", requesterID, "expires_at", data.ExpiresAt)
	audit.Log(ctx, audit.UserSuspend, audit.OnUser(user), audit.With("reason", reason), audit.With("expires_at", data.ExpiresAt))

	return ctx.JSON(populateUser(ctx, user))
}

func adminUnsuspendUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	var data types.UnsuspendUserRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	reason, err := suspensionReason(data.Reason)
	if err != nil {
		return err
	}

	lifted, err := star.Database.LiftSuspension(ctx, user.ID, requesterID, reason)
	if err != nil {
		return err
	} else if !lifted {
		return fiber.NewError(400, "This user is not suspended")
	}

	slog.Info("user unsuspended", "user", user.ID, "by", requesterID)
	audit.Log(ctx, audit.UserUnsuspend, audit.OnUser(user), audit.With("reason", reason))

	return ctx.JSON(populateUser(ctx, user))
}
