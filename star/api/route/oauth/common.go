package oauth

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
)

func getProperty(ctx *fiber.Ctx, key string) string {
	value := ctx.Query(key)
	if value != "" {
		return value
	}

	return ctx.FormValue(key)
}

func authorize(ctx *fiber.Ctx, provider *model.OAuthProvider) error {
	clientId := []byte(getProperty(ctx, "client_id"))

	if provider.ApplicationType == model.AppTypeConfidential {
		if helper.IsBasicAuth(ctx, provider.ClientID, provider.Secret) {
			return nil
		}

		clientSecret := []byte(getProperty(ctx, "client_secret"))
		if len(clientId) == 0 || len(clientSecret) == 0 {
			return ctx.Status(401).JSON(TokenError{
				Error:       TokenErrUnauthorizedClient,
				Description: "invalid application credentials provided",
			})
		}

		if subtle.ConstantTimeCompare(clientId, []byte(provider.ClientID)) != 1 ||
			subtle.ConstantTimeCompare(clientSecret, []byte(provider.Secret)) != 1 {
			return ctx.Status(401).JSON(TokenError{
				Error:       TokenErrUnauthorizedClient,
				Description: "invalid application credentials provided",
			})
		}

		return nil
	}

	if provider.ApplicationType == model.AppTypePublic {
		if len(clientId) == 0 || subtle.ConstantTimeCompare(clientId, []byte(provider.ClientID)) == 1 {
			return nil
		}

		return ctx.Status(401).JSON(TokenError{
			Error:       TokenErrUnauthorizedClient,
			Description: "invalid application credentials provided",
		})
	}

	return ctx.Status(401).JSON(TokenError{
		Error:       TokenErrServerError,
		Description: "failed to handle unknown application type: " + string(provider.ApplicationType),
	})
}
