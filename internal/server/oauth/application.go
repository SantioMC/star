package oauth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/server/common"
)

type PublicScope struct {
	Scope       string `json:"scope"`
	Description string `json:"description"`
	Application string `json:"application"`
}

type PublicApplication struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	URL             string        `json:"url"`
	DevelopmentMode bool          `json:"development_mode"`
	Scopes          []PublicScope `json:"scopes"`
	InvalidScopes   []string      `json:"invalid_scopes"`
}

func Application(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByClientID(ctx, ctx.Params("client_id"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	redirectURI := ctx.Query("redirect_uri")
	if redirectURI != "" && !AllowedRedirectURI(application, redirectURI) {
		return fiber.NewError(400, "The redirect_uri specified is not allowed for this application")
	}

	requested := strings.Fields(ctx.Query("scope"))
	if len(requested) == 0 {
		requested = []string{openid.ScopeOpenID}
	}

	return ctx.JSON(NewPublicApplication(ctx, application, requested))
}

func NewPublicApplication(ctx fiber.Ctx, application *database.Application, requested []string) PublicApplication {
	star := common.GetState(ctx)

	resolution := openid.ResolveScopes(ctx, star.Database, application, requested)
	scopes := make([]PublicScope, len(resolution.Granted))

	for idx, granted := range resolution.Granted {
		owner := granted.ApplicationName
		if owner == "" {
			owner = star.DefaultApp.Load().Name
		}

		scopes[idx] = PublicScope{
			Scope:       granted.Scope,
			Description: granted.Description,
			Application: owner,
		}
	}

	return PublicApplication{
		ID:              application.ID,
		Name:            application.Name,
		URL:             application.ApplicationURL,
		DevelopmentMode: application.DevelopmentMode,
		Scopes:          scopes,
		InvalidScopes:   resolution.Invalid,
	}
}
