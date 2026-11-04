package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
)

func alertRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminAlertsRead), middleware.RequirePermission(permission.PermissionAdminReadAlerts), getSecurityAlerts)
	router.Post("/:alert/resolve", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminAlertsWrite), middleware.RequirePermission(permission.PermissionAdminResolveAlerts), resolveSecurityAlert)
}

func getSecurityAlerts(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	page, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	unresolved := ctx.Query("status") == "open"
	return ctx.JSON(star.Database.GetSecurityAlerts(ctx, unresolved, page, limit))
}

func resolveSecurityAlert(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	alert := star.Database.GetSecurityAlertByID(ctx, ctx.Params("alert"))
	if alert == nil {
		return fiber.NewError(404, "Security alert not found")
	}

	resolved, err := star.Database.ResolveSecurityAlert(ctx, alert.ID, user.ID)
	if err != nil {
		return err
	} else if !resolved {
		return fiber.NewError(409, "This security alert has already been resolved")
	}

	audit.Log(ctx, audit.AlertResolve, audit.On(audit.TargetAlert, alert.ID, string(alert.Kind)))

	return ctx.JSON(star.Database.GetSecurityAlertByID(ctx, alert.ID))
}
