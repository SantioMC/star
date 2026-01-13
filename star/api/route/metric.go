package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func MetricRouter(r fiber.Router) {
	r.Get("/", getMetrics)
}

func getMetrics(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeReadMetrics); err != nil {
		return err
	}

	ctx.JSON(MetricsResponse{
		TotalUsers:    service.GetUserCount(),
		TotalSessions: service.GetActiveSessionCount(),
		StartedAt:     star.StartedAt.Unix(),
		NodeID:        star.NodeID,
		Version:       star.GetVersion(),
	})

	return nil
}

type MetricsResponse struct {
	TotalUsers    int64  `json:"total_users"`
	TotalSessions int64  `json:"total_sessions"`
	StartedAt     int64  `json:"started_at"`
	NodeID        string `json:"node_id"`
	Version       string `json:"version"`
}
