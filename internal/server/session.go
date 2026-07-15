package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
)

func sessionRouter(router fiber.Router) {
	router.Get("/", getUserSessions)
	router.Delete("/", middleware.RequirePermission(permission.PermissionDeleteSessions), invalidateUserSessions)

	router.Get("/:session", getUserSession)
	router.Delete("/:session", middleware.RequirePermission(permission.PermissionDeleteSessions), invalidateUserSession)
}

func getUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	sessions := star.Database.GetUserSessions(ctx, userId)
	return ctx.JSON(transformer.TransformSessions(sessions))
}

func invalidateUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	if err := star.Database.DeleteUserSessions(ctx, userId); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func getUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	session := star.Database.GetSessionByID(ctx, ctx.Params("session"))
	if session == nil || session.UserID != userId {
		return fiber.NewError(404, "session not found")
	}

	return ctx.JSON(transformer.TransformSession(*session))
}

func invalidateUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	session := star.Database.GetSessionByID(ctx, ctx.Params("session"))
	if session == nil || session.UserID != userId {
		return fiber.NewError(404, "session not found")
	}

	if err := star.Database.DeleteSessionByID(ctx, session.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
