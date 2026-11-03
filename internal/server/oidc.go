package server

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
)

func oidcRouter(router fiber.Router) {
	router.Get("/userinfo", middleware.RequireOAuth, getUserInfo)
	router.Post("/userinfo", middleware.RequireOAuth, getUserInfo)
}

func getUserInfo(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	firstParty := slices.Contains(jwt.Audience, star.DefaultApp.Load().ID)
	scopes := openid.EffectiveScopes(strings.Fields(jwt.Scope), firstParty)

	if !openid.HasScope(scopes, openid.ScopeOpenID) {
		ctx.Set(fiber.HeaderWWWAuthenticate, `Bearer error="insufficient_scope", error_description="The openid scope is required"`)
		return fiber.NewError(403, "The openid scope is required to access userinfo")
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(openid.UserClaims(ctx, star.Database, user, scopes))
}
