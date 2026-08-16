package oauth

import (
	"crypto/subtle"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
)

func authenticateClient(ctx fiber.Ctx, form url.Values) (*database.Application, bool, error) {
	star := common.GetState(ctx)

	clientId := form.Get("client_id")
	clientSecret := form.Get("client_secret")
	usedBasic := false

	header := ctx.Get(fiber.HeaderAuthorization)
	if header != "" {
		scheme, encoded, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "Basic") {
			return nil, false, invalidClient(ctx, false, "Unsupported client authentication scheme")
		}

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, false, invalidClient(ctx, true, "Malformed basic authentication header")
		}

		basicId, basicSecret, found := strings.Cut(string(decoded), ":")
		if !found {
			return nil, false, invalidClient(ctx, true, "Malformed basic authentication header")
		}

		if basicId, err = url.QueryUnescape(basicId); err != nil {
			return nil, false, invalidClient(ctx, true, "Malformed basic authentication header")
		}

		if basicSecret, err = url.QueryUnescape(basicSecret); err != nil {
			return nil, false, invalidClient(ctx, true, "Malformed basic authentication header")
		}

		if clientSecret != "" || (clientId != "" && clientId != basicId) {
			return nil, false, writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Multiple client authentication mechanisms were used")
		}

		clientId = basicId
		clientSecret = basicSecret
		usedBasic = true
	}

	if clientId == "" {
		return nil, true, nil
	}

	application := star.Database.GetApplicationByClientID(ctx, clientId)
	if application == nil {
		return nil, false, invalidClient(ctx, usedBasic, "Client authentication failed")
	}

	if application.Kind == database.ApplicationKindConfidential {
		if application.ClientSecret == nil || clientSecret == "" {
			return nil, false, invalidClient(ctx, usedBasic, "Client authentication failed")
		}

		if subtle.ConstantTimeCompare([]byte(*application.ClientSecret), []byte(clientSecret)) != 1 {
			return nil, false, invalidClient(ctx, usedBasic, "Client authentication failed")
		}
	} else if clientSecret != "" {
		return nil, false, invalidClient(ctx, usedBasic, "Client authentication failed")
	}

	return application, true, nil
}
