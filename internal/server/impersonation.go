package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/session"
)

func resolveImpersonation(ctx fiber.Ctx, impersonatorID string, userID string, application *database.Application) (*database.User, error) {
	star := common.GetState(ctx)

	if application.Internal {
		return nil, fiber.NewError(400, "Users can not be impersonated on this application")
	}

	user := star.Database.GetUserByID(ctx, userID)
	if user == nil {
		return nil, fiber.NewError(404, "User not found")
	}

	if err := session.CheckImpersonation(ctx, star, impersonatorID, user.ID); err != nil {
		return nil, err
	}

	if !star.Database.UserCanAccessApplication(ctx, impersonatorID, application) || !star.Database.UserCanAccessApplication(ctx, user.ID, application) {
		return nil, fiber.NewError(403, "You and this user do not share access to this application")
	}

	return user, nil
}

func getImpersonationApplications(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := session.CheckImpersonation(ctx, star, requesterID, user.ID); err != nil {
		return err
	}

	page, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	applications, err := star.Database.GetSharedApplications(ctx, requesterID, user.ID, ctx.Query("query"), page, limit)
	if err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformSharedApplications(applications))
}

func impersonateUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	var data types.ImpersonateUserRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	application := star.Database.GetApplicationByID(ctx, data.ApplicationID)
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	user, err := resolveImpersonation(ctx, jwt.Subject, ctx.Params("user"), application)
	if err != nil {
		return err
	}

	impersonation := &database.Impersonation{
		ID:             database.CreateID(database.ImpersonationKey),
		ImpersonatorID: jwt.Subject,
		UserID:         user.ID,
		ApplicationID:  application.ID,
		FamilyID:       jwt.SessionID,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Minute * 5),
	}

	if err := star.Database.SaveImpersonation(ctx, impersonation); err != nil {
		return err
	}

	audit.Log(ctx, audit.UserImpersonate, onApplication(application), audit.OnUser(user))

	return ctx.JSON(types.ImpersonateUserResponse{
		URL:       application.ApplicationURL,
		ExpiresAt: impersonation.ExpiresAt,
	})
}
