package oauth

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/service"
)

func HandleOidcToken(ctx *fiber.Ctx) error {
	ctx.Accepts("application/x-www-form-urlencoded")

	grantType := GrantType(getProperty(ctx, "grant_type"))
	if grantType == "" {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid form data provided, expected at least grant_type.",
		})
	}

	if !slices.Contains(KnownGrantTypes, grantType) {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidGrant,
			Description: "unknown grant type '" + string(grantType) + "'",
		})
	}

	switch grantType {
	case GrantAuthorizationCode:
		err := exchangeAuthorizationCode(ctx)
		if err != nil {
			slog.Error("failed to exchange authorization code", "err", err)
		}

		slog.Info("response data", "json", string(ctx.Response().Body()))

		return err
	case GrantRefreshToken:
		err := refreshAccessToken(ctx)
		if err != nil {
			slog.Error("failed to refresh access token", "err", err)
		}

		return err
	}

	return ctx.Status(500).JSON(TokenError{
		Error:       TokenErrServerError,
		Description: "star failed to properly process this request",
	})
}

func exchangeAuthorizationCode(ctx *fiber.Ctx) error {
	code := getProperty(ctx, "code")
	if code == "" {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid form data provided, expected code.",
		})
	}

	authRequest, err := service.GetAuthRequestByCode(code)
	if err != nil {
		return err
	}

	if authRequest == nil {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid authorization code provided",
		})
	}

	app := authRequest.Application
	provider := app.OauthProvider

	if provider == nil {
		slog.Error("failed to get oauth provider from application", "id", authRequest.ApplicationID, "app", app.ID)
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "the provided authorization code does not belong to a valid provider",
		})
	}

	if err := authorize(ctx, provider); err != nil {
		return err
	}

	if authRequest.IsExpired() {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "this authorization code has expired",
		})
	}

	if authRequest.CompletedAt != nil {
		// todo: report malicious behaviour
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "this authorization code has expired",
		})
	}

	if authRequest.RedirectURI == nil {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid callback uri provided",
		})
	}

	if authRequest.RedirectURI != nil {
		expectedRedirect := *authRequest.RedirectURI
		redirect := strings.TrimSpace(strings.ToLower(getProperty(ctx, "redirect_uri")))

		if redirect == "" || expectedRedirect != redirect {
			return ctx.Status(400).JSON(TokenError{
				Error:       TokenErrInvalidRequest,
				Description: "invalid callback uri provided",
			})
		}
	}

	if err := service.MarkAuthRequestCompleted(code); err != nil {
		return ctx.Status(500).JSON(TokenError{
			Error:       TokenErrServerError,
			Description: "failed to mark auth request as completed",
		})
	}

	session, jwt, expiresAt, err := service.CreateSession(
		authRequest.User,
		authRequest.Application,
		authRequest.Fingerprint,
		ctx.Get(fiber.HeaderUserAgent),
		ctx.IP(),
		oidc.SupportedClaims,
	)

	response := TokenResponse{
		AccessToken:  jwt,
		Expires:      expiresAt.Unix(),
		TokenType:    "Bearer",
		RefreshToken: &session.Token,
	}

	requestedScopes := authRequest.Scopes.Data()
	if !slices.Contains(requestedScopes, "offline_access") {
		response.RefreshToken = nil
	}

	return ctx.JSON(response)
}

func refreshAccessToken(ctx *fiber.Ctx) error {
	token := getProperty(ctx, "refresh_token")
	if token == "" {
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid form data provided, expected refresh_token.",
		})
	}

	session, err := service.GetSessionByToken(token)
	if err != nil {
		return err
	}

	if session == nil {
		slog.Warn("user attempted to refresh non-existent session")
		return ctx.Status(404).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid refresh token provided",
		})
	}

	app, err := service.GetAppByID(session.ApplicationID)
	if err != nil {
		slog.Error("failed to find app from session", "app_id", session.ApplicationID, "session", session.ID, "err", err)
		return ctx.Status(500).JSON(TokenError{
			Error:       TokenErrServerError,
			Description: "star failed to resolve app from session",
		})
	}

	provider := app.OauthProvider
	if provider == nil {
		slog.Warn("user attempted to refresh non-oauth2 session", "session", session.ID, "user", session.UserID)
		return ctx.Status(404).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "invalid refresh token provided",
		})
	}

	if err := authorize(ctx, provider); err != nil {
		return err
	}

	// If we detect an old token being used after we rotated it, we will flag it as possibly malicious
	if session.RotatedAt != nil {
		//todo: report malicious behaviour
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "this session has expired",
		})
	}

	if session.IsExpired() {
		slog.Warn("user attempted to refresh expired session", "session", session.ID, "user", session.UserID)
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: "this session has expired",
		})
	}

	// Time to refresh the token!
	fingerprint := getProperty(ctx, "fingerprint")
	userAgent := ctx.Get(fiber.HeaderUserAgent)

	newSession, jwt, expiresAt, err := service.RefreshSession(
		*session,
		&fingerprint,
		userAgent,
		ctx.IP(),
		provider.RotateRefreshToken,
	)

	if err != nil {
		slog.Warn("user failed to refresh access token", "session", session.ID, "user", session.UserID, "err", err)
		return ctx.Status(400).JSON(TokenError{
			Error:       TokenErrInvalidRequest,
			Description: err.Error(),
		})
	}

	var refreshToken *string = nil
	if newSession.Token != token {
		refreshToken = &newSession.Token
	}

	response := TokenResponse{
		AccessToken:  jwt,
		Expires:      expiresAt.Unix(),
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
	}

	return ctx.JSON(response)
}
