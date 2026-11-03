package oauth

import (
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/token"
)

type introspectionResponse struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
	Audience  []string `json:"aud,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	TokenID   string   `json:"jti,omitempty"`
	SessionID string   `json:"sid,omitempty"`
}

func Introspect(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	form, ok, err := parseForm(ctx)
	if !ok {
		return err
	}

	application, ok, err := authenticateClient(ctx, form)
	if !ok {
		return err
	}

	if application == nil || application.Kind != database.ApplicationKindConfidential {
		return invalidClient(ctx, false, "Only confidential clients may introspect tokens")
	}

	value := form.Get("token")
	if value == "" {
		return writeTokenError(ctx, 400, Oauth2ErrorInvalidRequest, "Missing token")
	}

	ctx.Set(fiber.HeaderCacheControl, "no-store")
	ctx.Set("Pragma", "no-cache")

	if strings.Count(value, ".") == 2 {
		return ctx.JSON(introspectAccessToken(star, application, value))
	}

	return ctx.JSON(introspectRefreshToken(ctx, star, application, value))
}

func introspectAccessToken(star *state.StarServer, application *database.Application, value string) introspectionResponse {
	jwt, err := token.Verify(star, value)
	if err != nil {
		return introspectionResponse{Active: false}
	}

	if jwt.ClientID != application.ClientID && !slices.Contains(jwt.Audience, application.ID) {
		return introspectionResponse{Active: false}
	}

	return introspectionResponse{
		Active:    true,
		Scope:     jwt.Scope,
		ClientID:  jwt.ClientID,
		TokenType: "Bearer",
		ExpiresAt: jwt.ExpiresAt,
		IssuedAt:  jwt.IssuedAt,
		NotBefore: jwt.NotBefore,
		Subject:   jwt.Subject,
		Audience:  jwt.Audience,
		Issuer:    jwt.Issuer,
		TokenID:   jwt.TokenID,
		SessionID: jwt.SessionID,
	}
}

func introspectRefreshToken(ctx fiber.Ctx, star *state.StarServer, application *database.Application, value string) introspectionResponse {
	session := star.Database.GetSessionByToken(ctx, value)
	if session == nil || session.ApplicationID != application.ID {
		return introspectionResponse{Active: false}
	}

	if session.Status != database.SessionStateActive || session.ExpiresAt.Before(time.Now()) {
		return introspectionResponse{Active: false}
	}

	return introspectionResponse{
		Active:    true,
		Scope:     strings.Join(session.Scopes, " "),
		ClientID:  application.ClientID,
		TokenType: "refresh_token",
		ExpiresAt: session.ExpiresAt.Unix(),
		IssuedAt:  session.CreatedAt.Unix(),
		Subject:   session.UserID,
		Issuer:    star.Config.ServerURL,
		SessionID: session.FamilyID,
	}
}
