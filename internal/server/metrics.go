package server

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/types"
)

func metricsRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminMetricsRead), middleware.RequirePermission(permission.PermissionAdminReadMetrics), getMetrics)
}

func getMetrics(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	users, err := star.Database.GetUserCount(ctx)
	if err != nil {
		return err
	}

	sessions, err := star.Database.GetSessionCount(ctx)
	if err != nil {
		return err
	}

	applications, err := star.Database.GetApplicationCount(ctx)
	if err != nil {
		return err
	}

	alerts, err := star.Database.CountUnresolvedSecurityAlerts(ctx)
	if err != nil {
		return err
	}

	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.Main.Version
	}

	return ctx.JSON(types.MetricsResponse{
		Users:          users,
		ActiveSessions: sessions,
		Applications:   applications,
		OpenAlerts:     alerts,
		Version:        version,
		StartedAt:      star.StartedAt,
	})
}
