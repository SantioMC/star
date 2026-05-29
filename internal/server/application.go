package server

import (
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/oauth"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
)

func appRouter(router fiber.Router) {
	router.Get("/", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionReadApplications), getApplications)
	router.Post("/", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionCreateApplications), createApplication)

	router.Patch("/:application", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionWriteApplications), updateApplication)
	router.Delete("/:application", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionDeleteApplications), deleteApplication)

	router.Post("/authorization", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionCreateAuthorization), createAppAuthorization)
}

func getApplications(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	return ctx.JSON(star.Database.GetApplications(ctx))
}

func createApplication(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.CreateApplicationRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	kind := database.ApplicationKind(request.Kind)
	if !slices.Contains(database.ApplicationKinds, kind) {
		return fiber.NewError(400, "Invalid application kind")
	}

	clientId := cryptography.CreateSecureString(24)
	if request.ClientID != nil {
		clientId = *request.ClientID
	}

	var clientSecret *string = nil
	if kind == database.ApplicationKindConfidential {
		secret := cryptography.CreateSecureString(24)
		clientSecret = &secret
	}

	application := &database.Application{
		ID:              database.CreateID(database.ApplicationKey),
		Name:            request.Name,
		ApplicationURL:  request.ApplicationURL,
		Kind:            kind,
		ClientID:        clientId,
		ClientSecret:    clientSecret,
		Internal:        false,
		TokenDuration:   models.Seconds(time.Hour),
		SessionDuration: models.Seconds(time.Hour * 24 * 7),
		RedirectURIs:    pq.StringArray{},
		DevelopmentMode: false,
		CreatedAt:       time.Now(),
	}

	if err := star.Database.CreateApplication(ctx, application); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func deleteApplication(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	applicationID := ctx.Params("application")

	application := star.Database.GetApplicationByID(ctx, applicationID)
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	if application.Internal {
		return fiber.NewError(401, "You are not permitted to delete this application")
	}

	if err := star.Database.DeleteApplicationByID(ctx, applicationID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func updateApplication(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	applicationID := ctx.Params("application")

	var request types.UpdateApplicationRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	application := star.Database.GetApplicationByID(ctx, applicationID)
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	if application.Internal {
		return fiber.NewError(401, "You are not permitted to modify this application")
	}

	if request.Name != nil {
		application.Name = *request.Name
	}

	if request.Kind != nil {
		kind := database.ApplicationKind(*request.Kind)
		if !slices.Contains(database.ApplicationKinds, kind) {
			return fiber.NewError(400, "Invalid application kind")
		}

		application.Kind = kind
	}

	if request.ApplicationURL != nil {
		application.ApplicationURL = *request.ApplicationURL
	}

	if request.ClientID != nil {
		application.ClientID = *request.ClientID
	}

	if request.TokenDuration != nil {
		application.TokenDuration = models.Seconds(*request.TokenDuration)
	}

	if request.SessionDuration != nil {
		application.SessionDuration = models.Seconds(*request.SessionDuration)
	}

	if application.SessionDuration <= application.TokenDuration {
		return fiber.NewError(400, "Session duration must be greater than token duration")
	}

	if request.RedirectURIs != nil {
		application.RedirectURIs = pq.StringArray(slices.Compact(request.RedirectURIs))
	}

	if request.DevelopmentMode != nil {
		application.DevelopmentMode = *request.DevelopmentMode
	}

	if err := star.Database.UpdateApplication(ctx, application); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func createAppAuthorization(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	var request types.CreateApplicationAuthorizationRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	application := star.Database.GetApplicationByID(ctx, request.ApplicationID)
	if application == nil {
		return fiber.NewError(400, "Application not found")
	}

	if !slices.Contains(application.RedirectURIs, request.RedirectURI) {
		return fiber.NewError(400, "Redirect URI is not allowed for this application")
	}

	scopes := strings.Split(request.Scope, " ")
	validScopes, modified := oauth.ValidateScopes(scopes)

	authCode := database.AuthorizationCode{
		ID:                  database.CreateID(database.AuthCodeKey),
		UserID:              user.ID,
		Application:         request.ApplicationID,
		Code:                cryptography.CreateSecureString(64),
		RedirectURI:         request.RedirectURI,
		Scopes:              validScopes,
		State:               request.State,
		Nonce:               request.Nonce,
		CodeChallenge:       request.CodeChallenge,
		CodeChallengeMethod: request.CodeChallengeMethod,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(time.Hour),
	}

	if err := star.Database.CreateAuthorizationCode(ctx, &authCode); err != nil {
		return err
	}

	return ctx.JSON(types.CreateApplicationAuthorizationResponse{
		Code:           authCode.Code,
		State:          authCode.State,
		Scopes:         validScopes,
		ScopedModified: modified,
	})
}
