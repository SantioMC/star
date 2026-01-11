package oauth

import (
	"log/slog"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
)

func HandleOidcToken(ctx *fiber.Ctx) error {
	ctx.Accepts("application/x-www-form-urlencoded")

	var request TokenRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "invalid form data provided, expected at least grant_type.",
		})
	}

	if request.GrantType == "" {
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "invalid form data provided, expected at least grant_type.",
		})
	}

	if !slices.Contains(KnownGrantTypes, request.GrantType) {
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidGrant,
			Description: "unknown grant type '" + string(request.GrantType) + "'",
		})
	}

	if request.GrantType == GrantAuthorizationCode {
		// todo: support authorization code
		return ctx.Status(500).JSON(TokenError{
			Error: TokenErrUnsupportedGrantType,
			Description: "this request is not made yet",
		})
	}

	if request.GrantType == GrantRefreshToken {
		var data RefreshTokenRequest
		if err := ctx.BodyParser(&data); err != nil {
			return ctx.Status(400).JSON(TokenError{
				Error: TokenErrInvalidRequest,
				Description: "invalid form data provided, expected both refresh_token and scope.",
			})
		}

		return refreshAccessToken(ctx, data)
	}

	return ctx.Status(500).JSON(TokenError{
		Error: TokenErrServerError,
		Description: "star failed to properly process this request",
	})
}

func refreshAccessToken(ctx *fiber.Ctx, data RefreshTokenRequest) error {
	session, err := service.GetSessionByToken(data.RefreshToken)
	if err != nil {
		return err
	}

	if session == nil {
		slog.Warn("user attempted to refresh non-existent session")
		return ctx.Status(404).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "invalid refresh token provided",
		})
	}

	app, err := service.GetAppByID(session.ApplicationID)
	if err != nil {
		slog.Error("failed to find app from session", "app_id", session.ApplicationID, "session", session.ID, "err", err)
		return ctx.Status(500).JSON(TokenError{
			Error: TokenErrServerError,
			Description: "star failed to resolve app from session",
		})
	}

	provider := app.OauthProvider
	if provider == nil {
		slog.Warn("user attempted to refresh non-oauth2 session", "session", session.ID, "user", session.UserID)
		return ctx.Status(404).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "invalid refresh token provided",
		})
	}

	if provider.ApplicationType == model.AppTypeConfidential {
		if !helper.IsBasicAuth(ctx, provider.ClientID, provider.Secret) {
			slog.Warn("user failed to authenticate with client", "client_id", provider.ClientID)
			return ctx.Status(401).JSON(TokenError{
				Error: TokenErrUnauthorizedClient,
				Description: "invalid application credentials provided",
			}) 
		}
	}

	// If we detect an old token being used after we rotated it, we will flag it as possibly malicious
	if session.RotatedAt != nil {
		//todo: report malicious behaviour
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "this session has expired",
		})
	}

	if session.IsExpired() {
		slog.Warn("user attempted to refresh expired session", "session", session.ID, "user", session.UserID)
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: "this session has expired",
		})
	}

	// Time to refresh the token!
	fingerprint := data.Fingerprint
	userAgent := ctx.Get(fiber.HeaderUserAgent)

	newSession, jwt, expiresAt, err := service.RefreshSession(
		*session,
		fingerprint,
		userAgent,
		provider.RotateRefreshToken,
	)

	if err != nil {
		slog.Warn("user failed to refresh access token", "session", session.ID, "user", session.UserID, "err", err)
		return ctx.Status(400).JSON(TokenError{
			Error: TokenErrInvalidRequest,
			Description: err.Error(),
		})
	}

	var refreshToken *string = nil
	if newSession.Token != data.RefreshToken {
		refreshToken = &newSession.Token
	}

	response := TokenResponse{
		AccessToken: jwt,
		Expires: expiresAt.Unix(),
		TokenType: "Bearer",
		RefreshToken: refreshToken,
	}

	return ctx.JSON(response)
}