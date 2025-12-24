package route

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func SessionRouter(r fiber.Router) {
	r.Get("/", getSessions)
	r.Delete("/", invalidateAllSessions)
	r.Get("/:session", getSession)
	r.Delete("/:session", invalidateSession)
}

func findSessions(userId string) []SessionResponse {
	sessions, err := service.GetAllSessions(userId)
	if err != nil {
		slog.Error("failed to get all sessions", "err", err)
		return []SessionResponse{}
	}

	response := []SessionResponse{}
	for _, session := range sessions {
		response = append(response, toSessionResponse(&session))
	}

	return response
}

func getSessions(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeReadSessions); err != nil {
		return err
	}

	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	sessions := findSessions(userId)
	return ctx.JSON(sessions)
}

func getSession(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeReadSessions); err != nil {
		return err
	}

	session, err := getSessionFromCtx(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(toSessionResponse(session))
}

func invalidateSession(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeDeleteSessions); err != nil {
		return err
	}

	session, err := getSessionFromCtx(ctx)
	if err != nil {
		return err
	}

	if err := service.InvalidateSession(session.ID); err != nil {
		return err
	}
	
	return ctx.JSON(fiber.Map{ "success": true })
}

func invalidateAllSessions(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeDeleteSessions); err != nil {
		return err
	}

	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	if err := service.InvalidateAllSessions(userId); err != nil {
		return err
	}
	
	return ctx.JSON(fiber.Map{ "success": true })
}

func getSessionFromCtx(ctx *fiber.Ctx) (*model.Session, error) {
	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return nil, err
	}

	sessionId := ctx.Params("session")
	session, err := service.GetSessionByID(sessionId)
	if err != nil {
		return nil, err
	}

	hasPermission := session.UserID == userId || helper.HasPermission(ctx, scopes.ScopeAdminReadUsers)
	if session == nil || !hasPermission {
		return nil, helper.NewHttpError(404, "session not found")
	}

	return session, nil
}

func toSessionResponse(session *model.Session) SessionResponse {
	return SessionResponse{
		ID:     session.ID,
		Device: session.GetDevice(),
		OS:     session.GetOS(),
		IP:     session.Address,

		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt(),
		Expired:   session.IsExpired(),
	}
}

type SessionResponse struct {
	ID        string `json:"id"`
	Device    string `json:"device"`
	OS        string `json:"operating_system"`
	IP        string `json:"ip_address"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	Expired   bool   `json:"expired"`
}
