package server

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/oauth"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func deviceRouter(router fiber.Router) {
	router.Use(
		middleware.RateLimit(10),
		middleware.RequireAuth,
		middleware.RequireScope(openid.ScopeUserAuthorizeApps),
		middleware.RequirePermission(permission.PermissionCreateAuthorization),
		middleware.UserFailureLimit(5, time.Minute*10),
	)

	router.Post("/verify", verifyDevice)
	router.Post("/decision", decideDevice)
}

func pendingDeviceAuthorization(ctx fiber.Ctx, star *state.StarServer, userCode string, userID string) (*database.DeviceAuthorization, *database.Application, error) {
	authorization := star.Database.GetPendingDeviceAuthorization(ctx, oauth.NormalizeUserCode(userCode))
	if authorization == nil {
		return nil, nil, fiber.NewError(404, "This code is invalid or has expired, check the code on your device and try again")
	}

	application := star.Database.GetApplicationByID(ctx, authorization.ApplicationID)
	if application == nil || !application.DeviceFlow {
		return nil, nil, fiber.NewError(404, "This code is invalid or has expired, check the code on your device and try again")
	}

	if !star.Database.UserCanAccessApplication(ctx, userID, application) {
		return nil, nil, fiber.NewError(403, "application_access_denied")
	}

	return authorization, application, nil
}

func verifyDevice(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	var body types.VerifyDeviceRequest
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	authorization, application, err := pendingDeviceAuthorization(ctx, star, body.UserCode, userID)
	if err != nil {
		return err
	}

	ctx.Set(fiber.HeaderCacheControl, "no-store")
	return ctx.JSON(types.VerifyDeviceResponse{
		UserCode:    oauth.FormatUserCode(oauth.NormalizeUserCode(body.UserCode)),
		Device:      authorization.Device,
		Geolocation: authorization.Geolocation,
		ExpiresAt:   authorization.ExpiresAt,
		Application: oauth.NewPublicApplication(ctx, application, authorization.Scopes),
	})
}

func decideDevice(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	var body types.DecideDeviceRequest
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	authorization, application, err := pendingDeviceAuthorization(ctx, star, body.UserCode, user.ID)
	if err != nil {
		return err
	}

	if !body.Approve {
		return denyDevice(ctx, star, authorization, application, user)
	}

	return approveDevice(ctx, star, authorization, application, user)
}

func denyDevice(ctx fiber.Ctx, star *state.StarServer, authorization *database.DeviceAuthorization, application *database.Application, user *database.User) error {
	denied, err := star.Database.DenyDeviceAuthorization(ctx, authorization.ID, user.ID)
	if err != nil {
		return err
	} else if !denied {
		return fiber.NewError(409, "This code has already been used")
	}

	audit.Log(ctx, audit.DeviceDeny, onApplication(application))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func approveDevice(ctx fiber.Ctx, star *state.StarServer, authorization *database.DeviceAuthorization, application *database.Application, user *database.User) error {
	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	resolution := openid.ResolveScopes(ctx, star.Database, application, authorization.Scopes)
	if len(resolution.Granted) == 0 {
		return fiber.NewError(400, "None of the requested scopes are supported")
	}

	scopes := resolution.Names()
	grant := star.Database.GetApplicationGrant(ctx, user.ID, application.ID)

	if err := saveApplicationGrant(ctx, star, application, user, grant, scopes); err != nil {
		return err
	}

	authenticatedAt := sessionAuthenticatedAt(ctx, star, jwt.SessionID, user.ID)

	approved, err := star.Database.ApproveDeviceAuthorization(ctx, authorization.ID, user.ID, scopes, authenticatedAt)
	if err != nil {
		return err
	} else if !approved {
		return fiber.NewError(409, "This code has already been used")
	}

	audit.Log(ctx, audit.DeviceApprove, onApplication(application), audit.With("scopes", scopes))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
