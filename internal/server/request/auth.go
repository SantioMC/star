package request

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
)

func HasScope(ctx fiber.Ctx, scope string) bool {
	star := common.GetState(ctx)

	jwt, err := common.RequestJWT(ctx)
	if err != nil {
		return false
	}

	firstParty := slices.Contains(jwt.Audience, star.DefaultApp.Load().ID)
	if openid.IsStarScope(scope) && !firstParty {
		return false
	}

	granted := openid.EffectiveScopes(strings.Fields(jwt.Scope), firstParty)
	return openid.HasScope(granted, scope)
}
