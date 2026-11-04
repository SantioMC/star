package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/keypair"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/transformer"
)

func keyRouter(router fiber.Router) {
	readKeys := middleware.RequireScope(openid.ScopeAdminKeysRead)
	writeKeys := middleware.RequireScope(openid.ScopeAdminKeysWrite)

	router.Get("/", middleware.RequireAuth, readKeys, middleware.RequirePermission(permission.PermissionAdminReadKeys), getSigningKeys)
	router.Post("/rotate", middleware.RequireAuth, writeKeys, middleware.RequirePermission(permission.PermissionAdminRotateKeys), rotateSigningKeys)
	router.Delete("/:key", middleware.RequireAuth, writeKeys, middleware.RequirePermission(permission.PermissionAdminDeleteKeys), deleteSigningKey)
}

func getSigningKeys(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	keys, err := transformer.TransformSigningKeys(star.Database.GetAllSigningKeys(ctx))
	if err != nil {
		return err
	}

	return ctx.JSON(keys)
}

func rotateSigningKeys(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	if err := keypair.Rotate(ctx, star); err != nil {
		return err
	}

	cache.UpdateKeys(ctx, star)
	audit.Log(ctx, audit.KeyRotate)

	return getSigningKeys(ctx)
}

func deleteSigningKey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	key := star.Database.GetSigningKeyByID(ctx, ctx.Params("key"))
	if key == nil {
		return fiber.NewError(404, "Signing key not found")
	}

	deleted, err := star.Database.DeleteRetiringSigningKey(ctx, key.ID)
	if err != nil {
		return err
	} else if !deleted {
		return fiber.NewError(409, "Only retiring signing keys can be deleted")
	}

	cache.UpdateKeys(ctx, star)
	audit.Log(ctx, audit.KeyDelete, audit.On(audit.TargetKey, key.ID, ""))

	return getSigningKeys(ctx)
}
