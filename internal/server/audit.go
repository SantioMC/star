package server

import (
	"net/netip"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
)

func auditRouter(router fiber.Router) {
	read := middleware.RequireScope(openid.ScopeAdminAuditRead)
	readPermission := middleware.RequirePermission(permission.PermissionAdminReadAudit)

	router.Get("/", middleware.RequireAuth, read, readPermission, getAuditLogs)
	router.Get("/actions", middleware.RequireAuth, read, readPermission, getAuditActions)
	router.Get("/settings", middleware.RequireAuth, read, readPermission, getAuditSettings)
	router.Patch("/settings", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminAuditWrite), middleware.RequirePermission(permission.PermissionAdminWriteAudit), updateAuditSettings)
}

func timeQuery(ctx fiber.Ctx, key string) (*time.Time, error) {
	value := ctx.Query(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fiber.NewError(400, key+" must be an RFC 3339 timestamp")
	}

	return &parsed, nil
}

func applyAuditSearch(ctx fiber.Ctx, filter *database.AuditLogFilter, search string) error {
	star := common.GetState(ctx)

	if address, err := netip.ParseAddr(search); err == nil {
		filter.AddressKey = new(star.Database.HashAuditAddress(address.Unmap().String()))
		return nil
	}

	matches, err := star.Database.FindUserIDs(ctx, search, 20)
	if err != nil {
		return err
	}

	filter.Subjects = append(matches, search)
	return nil
}

func parseAuditFilter(ctx fiber.Ctx) (*database.AuditLogFilter, error) {
	_, limit, err := request.GetPagination(ctx)
	if err != nil {
		return nil, err
	}

	filter := &database.AuditLogFilter{Limit: limit + 1}

	if user := ctx.Query("user"); user != "" {
		filter.UserID = &user
	}

	if application := ctx.Query("application"); application != "" {
		filter.Application = &application
	}

	if provider := ctx.Query("provider"); provider != "" {
		filter.Provider = &provider
	}

	if action := ctx.Query("action"); action != "" {
		filter.Actions = []string{action}
	} else if category := ctx.Query("category"); category != "" {
		filter.Actions = audit.ActionsInCategory(category)
		if len(filter.Actions) == 0 {
			return nil, fiber.NewError(400, "Unknown audit log category")
		}
	}

	if filter.Since, err = timeQuery(ctx, "since"); err != nil {
		return nil, err
	}

	if filter.Until, err = timeQuery(ctx, "until"); err != nil {
		return nil, err
	}

	if filter.Before, err = request.GetCursor(ctx); err != nil {
		return nil, err
	}

	if search := strings.TrimSpace(ctx.Query("query")); search != "" {
		if err := applyAuditSearch(ctx, filter, search); err != nil {
			return nil, err
		}
	}

	return filter, nil
}

func getAuditLogs(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	filter, err := parseAuditFilter(ctx)
	if err != nil {
		return err
	}

	logs, err := star.Database.GetAuditLogs(ctx, *filter)
	if err != nil {
		return err
	}

	size := filter.Limit - 1
	page := types.NewPage(logs, size, func(last database.AuditLog) database.Cursor {
		return database.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	})

	return ctx.JSON(page)
}

func getAuditActions(ctx fiber.Ctx) error {
	return ctx.JSON(audit.Definitions)
}

func getAuditSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(transformer.TransformAuditSettings(*settings))
}

func updateAuditSettings(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.UpdateAuditSettingsRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	settings, err := currentSettings(star, ctx)
	if err != nil {
		return err
	}

	before := transformer.TransformAuditSettings(*settings)

	if data.RetentionDays != nil {
		days := *data.RetentionDays

		if days != 0 && days < 7 {
			return fiber.NewError(400, "Audit logs must be retained for at least 7 days")
		}

		settings.AuditRetentionDays = nil
		if days != 0 {
			settings.AuditRetentionDays = &days
		}
	}

	if data.Anonymize != nil {
		settings.AuditAnonymize = *data.Anonymize
	}

	if err := star.Database.UpdateSettings(ctx, settings); err != nil {
		return err
	}

	cache.UpdateSettings(ctx, star)

	after := transformer.TransformAuditSettings(*settings)
	audit.Log(ctx, audit.AuditSettingsUpdate, audit.With("before", before), audit.With("after", after))

	return ctx.JSON(after)
}
