package server

import (
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/oauth"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func appRouter(router fiber.Router) {
	readApplications := middleware.RequireScope(openid.ScopeApplicationsRead)
	writeApplications := middleware.RequireScope(openid.ScopeApplicationsWrite)
	deleteApplications := middleware.RequireScope(openid.ScopeApplicationsDelete)

	router.Get("/", readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplications)
	router.Post("/", writeApplications, middleware.RequirePermission(permission.PermissionCreateApplications), createApplication)

	router.Post(
		"/authorization",
		middleware.RequireScope(openid.ScopeUserAuthorizeApps),
		middleware.RequirePermission(permission.PermissionCreateAuthorization),
		createAppAuthorization,
	)

	router.Get("/:application", readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplication)
	router.Patch("/:application", writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), updateApplication)
	router.Delete("/:application", deleteApplications, middleware.RequirePermission(permission.PermissionDeleteApplications), deleteApplication)

	router.Post("/:application/rotate", writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), rotateApplicationSecret)

	router.Get("/:application/scopes", readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplicationScopes)
	router.Post("/:application/scopes", writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), createApplicationScope)
	router.Patch("/:application/scopes/:scope", writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), updateApplicationScope)
	router.Delete("/:application/scopes/:scope", writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), deleteApplicationScope)
}

func getApplications(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	applications := star.Database.GetApplications(ctx)
	scopes := star.Database.GetAllApplicationScopes(ctx)

	return ctx.JSON(transformer.TransformApplications(applications, scopes))
}

func getApplication(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByID(ctx, ctx.Params("application"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	scopes := star.Database.GetApplicationScopes(ctx, application.ID)
	return ctx.JSON(transformer.TransformApplication(*application, scopes))
}

func rotateApplicationSecret(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByID(ctx, ctx.Params("application"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	if application.Internal {
		return fiber.NewError(403, "You are not permitted to modify this application")
	}

	if application.Kind != database.ApplicationKindConfidential {
		return fiber.NewError(400, "Only confidential applications have a client secret")
	}

	secret := database.CreateApplicationSecret()
	application.ClientSecret = &secret

	if err := star.Database.UpdateApplication(ctx, application); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"client_secret": secret,
	})
}

func resolveNamespace(ctx fiber.Ctx, star *state.StarServer, requested string, fallback string) (string, error) {
	namespace := openid.NormalizeNamespace(requested)
	if namespace == "" {
		namespace = openid.NormalizeNamespace(fallback)
	}

	if !openid.ValidNamespace(namespace) {
		return "", fiber.NewError(400, "The namespace must be 2-32 characters of lowercase letters, digits, and dashes, starting with a letter")
	}

	if openid.ReservedNamespace(namespace) {
		return "", fiber.NewError(400, "The namespace '"+namespace+"' is reserved")
	}

	if existing := star.Database.GetApplicationByNamespace(ctx, namespace); existing != nil {
		return "", fiber.NewError(400, "Another application already uses the namespace '"+namespace+"'")
	}

	return namespace, nil
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

	requestedNamespace := ""
	if request.Namespace != nil {
		requestedNamespace = *request.Namespace
	}

	namespace, err := resolveNamespace(ctx, star, requestedNamespace, request.Name)
	if err != nil {
		return err
	}

	clientId := cryptography.CreateSecureString(24)
	if request.ClientID != nil {
		clientId = *request.ClientID
	}

	if star.Database.GetApplicationByClientID(ctx, clientId) != nil {
		return fiber.NewError(400, "Another application already uses this client id")
	}

	var clientSecret *string = nil
	if kind == database.ApplicationKindConfidential {
		secret := database.CreateApplicationSecret()
		clientSecret = &secret
	}

	application := &database.Application{
		ID:              database.CreateID(database.ApplicationKey),
		Name:            request.Name,
		Namespace:       namespace,
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

	if request.Namespace != nil {
		namespace := openid.NormalizeNamespace(*request.Namespace)

		if namespace != application.Namespace {
			if len(star.Database.GetApplicationScopes(ctx, application.ID)) > 0 {
				return fiber.NewError(400, "Delete this application's scopes before changing its namespace")
			}

			resolved, err := resolveNamespace(ctx, star, namespace, application.Name)
			if err != nil {
				return err
			}

			application.Namespace = resolved
		}
	}

	if request.Kind != nil {
		kind := database.ApplicationKind(*request.Kind)
		if !slices.Contains(database.ApplicationKinds, kind) {
			return fiber.NewError(400, "Invalid application kind")
		}

		if kind != application.Kind {
			application.Kind = kind

			if kind == database.ApplicationKindConfidential {
				secret := database.CreateApplicationSecret()
				application.ClientSecret = &secret
			} else {
				application.ClientSecret = nil
			}
		}
	}

	if request.ApplicationURL != nil {
		application.ApplicationURL = *request.ApplicationURL
	}

	if request.ClientID != nil && *request.ClientID != application.ClientID {
		if star.Database.GetApplicationByClientID(ctx, *request.ClientID) != nil {
			return fiber.NewError(400, "Another application already uses this client id")
		}

		application.ClientID = *request.ClientID
	}

	if request.TokenDuration != nil {
		application.TokenDuration = models.Seconds(*request.TokenDuration)
	}

	if request.SessionDuration != nil {
		application.SessionDuration = models.Seconds(*request.SessionDuration)
	}

	if application.TokenDuration.Duration() > time.Hour*24*3 {
		return fiber.NewError(400, "Token duration may not exceed 3 days")
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

func modifiableApplication(ctx fiber.Ctx) (*database.Application, error) {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByID(ctx, ctx.Params("application"))
	if application == nil {
		return nil, fiber.NewError(404, "Application not found")
	}

	if application.Internal {
		return nil, fiber.NewError(403, "You are not permitted to modify this application")
	}

	return application, nil
}

func getApplicationScopes(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByID(ctx, ctx.Params("application"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	return ctx.JSON(star.Database.GetApplicationScopes(ctx, application.ID))
}

func validateScopeName(star *state.StarServer, ctx fiber.Ctx, application *database.Application, name string) error {
	trimmed := strings.TrimSpace(name)

	if !openid.ValidScopeName(application.Namespace, trimmed) {
		return fiber.NewError(400, "Scopes must be named '"+application.Namespace+":<key>' using lowercase letters, digits, dots, dashes, underscores, and colons")
	}

	if openid.Builtin(trimmed) != nil {
		return fiber.NewError(400, "This scope name is reserved by star")
	}

	if existing := star.Database.GetApplicationScopeByName(ctx, trimmed); existing != nil {
		return fiber.NewError(400, "A scope with this name already exists")
	}

	return nil
}

func createApplicationScope(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application, err := modifiableApplication(ctx)
	if err != nil {
		return err
	}

	var request types.CreateApplicationScopeRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	name := strings.TrimSpace(request.Name)
	if err := validateScopeName(star, ctx, application, name); err != nil {
		return err
	}

	scope := &database.ApplicationScope{
		ID:            database.CreateID(database.ApplicationScopeKey),
		ApplicationID: application.ID,
		Name:          name,
		Description:   strings.TrimSpace(request.Description),
		Requestable:   request.Requestable,
		CreatedAt:     time.Now(),
	}

	if err := star.Database.CreateApplicationScope(ctx, scope); err != nil {
		return err
	}

	return ctx.JSON(scope)
}

func updateApplicationScope(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application, err := modifiableApplication(ctx)
	if err != nil {
		return err
	}

	scope := star.Database.GetApplicationScopeByID(ctx, ctx.Params("scope"))
	if scope == nil || scope.ApplicationID != application.ID {
		return fiber.NewError(404, "Scope not found")
	}

	var request types.UpdateApplicationScopeRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)

		if name != scope.Name {
			if err := validateScopeName(star, ctx, application, name); err != nil {
				return err
			}

			scope.Name = name
		}
	}

	if request.Description != nil {
		scope.Description = strings.TrimSpace(*request.Description)
	}

	if request.Requestable != nil {
		scope.Requestable = *request.Requestable
	}

	if err := star.Database.UpdateApplicationScope(ctx, scope); err != nil {
		return err
	}

	return ctx.JSON(scope)
}

func deleteApplicationScope(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application, err := modifiableApplication(ctx)
	if err != nil {
		return err
	}

	scope := star.Database.GetApplicationScopeByID(ctx, ctx.Params("scope"))
	if scope == nil || scope.ApplicationID != application.ID {
		return fiber.NewError(404, "Scope not found")
	}

	if err := star.Database.DeleteApplicationScopeByID(ctx, scope.ID); err != nil {
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

	jwt, err := request.GetJWT(ctx)
	if err != nil {
		return err
	}

	authenticatedAt := time.Now()
	if current := star.Database.GetSessionByID(ctx, jwt.SessionID); current != nil {
		authenticatedAt = current.AuthenticatedAt
	} else {
		slog.Warn("issuing an authorization code for a session that no longer exists", "session", jwt.SessionID, "user", user.ID)
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

	challenge := ""
	if request.CodeChallenge != nil {
		challenge = *request.CodeChallenge
	}

	method := ""
	if request.CodeChallengeMethod != nil {
		method = *request.CodeChallengeMethod
	}

	if _, description, ok := oauth.ValidateCodeChallenge(application, challenge, method); !ok {
		return fiber.NewError(400, description)
	}

	scopes := strings.Fields(request.Scope)
	resolution := openid.ResolveScopes(ctx, star.Database, application, scopes)

	if len(resolution.Granted) == 0 {
		return fiber.NewError(400, "None of the requested scopes are supported")
	}

	validScopes := resolution.Names()

	var codeChallenge *string = nil
	var codeChallengeMethod *string = nil

	if challenge != "" {
		s256 := oauth.CodeChallengeMethodS256
		codeChallenge = &challenge
		codeChallengeMethod = &s256
	}

	authCode := database.AuthorizationCode{
		ID:                  database.CreateID(database.AuthCodeKey),
		UserID:              user.ID,
		ApplicationID:       request.ApplicationID,
		Code:                cryptography.CreateSecureString(64),
		RedirectURI:         request.RedirectURI,
		Scopes:              pq.StringArray(validScopes),
		State:               request.State,
		Nonce:               request.Nonce,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		AuthenticatedAt:     authenticatedAt,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(time.Minute * 5),
	}

	if err := star.Database.CreateAuthorizationCode(ctx, &authCode); err != nil {
		return err
	}

	return ctx.JSON(types.CreateApplicationAuthorizationResponse{
		Code:           authCode.Code,
		State:          authCode.State,
		Scopes:         validScopes,
		ScopedModified: len(resolution.Invalid) > 0,
	})
}
