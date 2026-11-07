package server

import (
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/cache"
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

	router.Get("/", middleware.RequireAuth, readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplications)
	router.Post("/", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionCreateApplications), createApplication)

	router.Post(
		"/authorization",
		middleware.RequireAuth,
		middleware.RequireScope(openid.ScopeUserAuthorizeApps),
		middleware.RequirePermission(permission.PermissionCreateAuthorization),
		createAppAuthorization,
	)

	router.Get("/:application", middleware.RequireAuth, readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplication)
	router.Patch("/:application", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), updateApplication)
	router.Delete("/:application", middleware.RequireAuth, deleteApplications, middleware.RequirePermission(permission.PermissionDeleteApplications), deleteApplication)

	router.Post("/:application/rotate", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), rotateApplicationSecret)

	router.Get("/:application/scopes", middleware.RequireAuth, readApplications, middleware.RequirePermission(permission.PermissionReadApplications), getApplicationScopes)
	router.Post("/:application/scopes", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), createApplicationScope)
	router.Patch("/:application/scopes/:scope", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), updateApplicationScope)
	router.Delete("/:application/scopes/:scope", middleware.RequireAuth, writeApplications, middleware.RequirePermission(permission.PermissionWriteApplications), deleteApplicationScope)
}

func getApplications(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	applications := star.Database.GetApplications(ctx)
	scopes := star.Database.GetAllApplicationScopes(ctx)
	bindings := star.Database.GetAllApplicationRoleBindings(ctx)

	return ctx.JSON(transformer.TransformApplications(applications, scopes, bindings))
}

func getApplication(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	application := star.Database.GetApplicationByID(ctx, ctx.Params("application"))
	if application == nil {
		return fiber.NewError(404, "Application not found")
	}

	scopes := star.Database.GetApplicationScopes(ctx, application.ID)
	roles := star.Database.GetApplicationRoleIDs(ctx, application.ID)

	return ctx.JSON(transformer.TransformApplication(*application, scopes, roles))
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

	secret := assignClientSecret(star, application)
	if err := star.Database.UpdateApplication(ctx, application); err != nil {
		return err
	}

	audit.Log(ctx, audit.ApplicationSecretRotate, onApplication(application))

	return ctx.JSON(fiber.Map{
		"client_secret": secret,
	})
}

func onApplication(application *database.Application) audit.Option {
	return audit.On(audit.TargetApplication, application.ID, application.Name)
}

func assignClientSecret(star *state.StarServer, application *database.Application) string {
	secret := database.CreateApplicationSecret()
	hashed := star.Database.HashApplicationSecret(secret)
	application.ClientSecret = &hashed

	return secret
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

	if err := oauth.ValidateApplicationURL(request.ApplicationURL); err != nil {
		return fiber.NewError(400, err.Error())
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
		if err := oauth.ValidateClientID(*request.ClientID); err != nil {
			return fiber.NewError(400, err.Error())
		}

		clientId = *request.ClientID
	}

	if star.Database.GetApplicationByClientID(ctx, clientId) != nil {
		return fiber.NewError(400, "Another application already uses this client id")
	}

	application := &database.Application{
		ID:               database.CreateID(database.ApplicationKey),
		Name:             request.Name,
		Namespace:        namespace,
		ApplicationURL:   request.ApplicationURL,
		Kind:             kind,
		ClientID:         clientId,
		Internal:         false,
		Consent:          database.ApplicationConsentDefault,
		TokenDuration:    models.Seconds(time.Hour),
		SessionDuration:  models.Seconds(time.Hour * 24 * 7),
		RedirectURIs:     pq.StringArray{},
		DevelopmentMode:  false,
		SigningAlgorithm: database.KeyAlgorithmES512,
		CreatedAt:        time.Now(),
	}

	var clientSecret *string
	if kind == database.ApplicationKindConfidential {
		clientSecret = new(assignClientSecret(star, application))
	}

	if err := star.Database.CreateApplication(ctx, application); err != nil {
		return err
	}

	audit.Log(ctx, audit.ApplicationCreate, onApplication(application))

	return ctx.JSON(types.ApplicationSecretResponse{
		ID:           application.ID,
		ClientSecret: clientSecret,
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

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.ApplicationDelete, onApplication(application))

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

	before := *application

	var clientSecret *string

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
				clientSecret = new(assignClientSecret(star, application))
			} else {
				application.ClientSecret = nil
			}
		}
	}

	if request.Consent != nil {
		consent := database.ApplicationConsent(*request.Consent)
		if !slices.Contains(database.ApplicationConsents, consent) {
			return fiber.NewError(400, "Invalid consent mode")
		}

		application.Consent = consent
	}

	if request.ApplicationURL != nil {
		if err := oauth.ValidateApplicationURL(*request.ApplicationURL); err != nil {
			return fiber.NewError(400, err.Error())
		}

		application.ApplicationURL = *request.ApplicationURL
	}

	if request.ClientID != nil && *request.ClientID != application.ClientID {
		if err := oauth.ValidateClientID(*request.ClientID); err != nil {
			return fiber.NewError(400, err.Error())
		}

		if star.Database.GetApplicationByClientID(ctx, *request.ClientID) != nil {
			return fiber.NewError(400, "Another application already uses this client id")
		}

		application.ClientID = *request.ClientID
	}

	if request.TokenDuration != nil {
		application.TokenDuration = *request.TokenDuration
	}

	if request.SessionDuration != nil {
		application.SessionDuration = *request.SessionDuration
	}

	tokenDuration := application.TokenDuration.Duration()
	if tokenDuration < time.Minute || tokenDuration > time.Hour {
		return fiber.NewError(400, "Token duration must be between 1 minute and 1 hour")
	}

	if application.SessionDuration <= application.TokenDuration {
		return fiber.NewError(400, "Session duration must be greater than token duration")
	}

	if request.RedirectURIs != nil {
		for _, uri := range request.RedirectURIs {
			if err := oauth.ValidateRedirectURI(uri); err != nil {
				return fiber.NewError(400, err.Error())
			}
		}

		application.RedirectURIs = pq.StringArray(slices.Compact(request.RedirectURIs))
	}

	if request.DevelopmentMode != nil {
		application.DevelopmentMode = *request.DevelopmentMode
	}

	if request.DeviceFlow != nil {
		application.DeviceFlow = *request.DeviceFlow
	}

	if request.PrivilegedScopes != nil {
		application.PrivilegedScopes = *request.PrivilegedScopes
	}

	if request.SigningAlgorithm != nil {
		algorithm := database.KeyAlgorithm(*request.SigningAlgorithm)
		if !slices.Contains(database.KeyAlgorithms, algorithm) {
			return fiber.NewError(400, "Invalid signing algorithm")
		}

		application.SigningAlgorithm = algorithm
	}

	if request.RoleRestricted != nil {
		application.RoleRestricted = *request.RoleRestricted
	}

	options := []audit.Option{onApplication(application), audit.Changes(before, *application)}

	if request.RoleRestricted != nil || request.Roles != nil {
		roles, err := resolveApplicationRoles(ctx, star, application, request.Roles)
		if err != nil {
			return err
		}

		if err := star.Database.UpdateApplicationWithRoles(ctx, application, roles); err != nil {
			return err
		}

		cache.UpdateRevocations(ctx, star)
		options = append(options, audit.With("roles", roles))
	} else if err := star.Database.UpdateApplication(ctx, application); err != nil {
		return err
	}

	audit.Log(ctx, audit.ApplicationUpdate, options...)

	return ctx.JSON(types.ApplicationSecretResponse{
		ID:           application.ID,
		ClientSecret: clientSecret,
	})
}

func resolveApplicationRoles(
	ctx fiber.Ctx,
	star *state.StarServer,
	application *database.Application,
	requested []string,
) ([]string, error) {
	if requested == nil {
		requested = star.Database.GetApplicationRoleIDs(ctx, application.ID)
	}

	roles := slices.Compact(slices.Sorted(slices.Values(requested)))

	for _, id := range roles {
		role := star.Database.GetRoleByID(ctx, id)
		if role == nil {
			return nil, fiber.NewError(400, "Unknown role: "+id)
		}

		if role.IsDefault() {
			return nil, fiber.NewError(400, "The default role applies to every user and can not restrict an application")
		}
	}

	if application.RoleRestricted && len(roles) == 0 {
		return nil, fiber.NewError(400, "At least one role is required to restrict an application")
	}

	return roles, nil
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

	audit.Log(ctx, audit.ApplicationScopeCreate, onApplication(application), audit.With("scope", scope.Name))

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

	audit.Log(ctx, audit.ApplicationScopeUpdate, onApplication(application), audit.With("scope", scope.Name))

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

	audit.Log(ctx, audit.ApplicationScopeDelete, onApplication(application), audit.With("scope", scope.Name))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func sessionAuthenticatedAt(ctx fiber.Ctx, star *state.StarServer, sessionID string, userID string) time.Time {
	current := star.Database.GetSessionByFamily(ctx, sessionID)
	if current == nil {
		slog.Warn("issuing an authorization for a session that no longer exists", "session", sessionID, "user", userID)
		return time.Now()
	}

	return current.AuthenticatedAt
}

func saveApplicationGrant(ctx fiber.Ctx, star *state.StarServer, application *database.Application, user *database.User, grant *database.ApplicationGrant, scopes []string) error {
	merged := oauth.MergeGrantScopes(grant, scopes)

	err := star.Database.SaveApplicationGrant(ctx, &database.ApplicationGrant{
		ID:            database.CreateID(database.ApplicationGrantKey),
		UserID:        user.ID,
		ApplicationID: application.ID,
		Scopes:        pq.StringArray(merged),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	if err != nil {
		return err
	}

	if grant == nil || len(merged) != len(grant.Scopes) {
		audit.Log(ctx, audit.GrantCreate, onApplication(application), audit.With("scopes", merged))
	}

	return nil
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

	authenticatedAt := sessionAuthenticatedAt(ctx, star, jwt.SessionID, user.ID)

	var request types.CreateApplicationAuthorizationRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	application := star.Database.GetApplicationByID(ctx, request.ApplicationID)
	if application == nil {
		return fiber.NewError(400, "Application not found")
	}

	if !oauth.AllowedRedirectURI(application, request.RedirectURI) {
		return fiber.NewError(400, "Redirect URI is not allowed for this application")
	}

	subject := user
	impersonation := star.Database.GetImpersonation(ctx, user.ID, jwt.SessionID, application.ID)

	if impersonation != nil {
		subject, err = resolveImpersonation(ctx, user.ID, impersonation.UserID, application)
		if err != nil {
			return err
		}
	}

	if !star.Database.UserCanAccessApplication(ctx, subject.ID, application) {
		return fiber.NewError(403, "application_access_denied")
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
	grant := star.Database.GetApplicationGrant(ctx, user.ID, application.ID)

	if impersonation == nil && !request.Consent && oauth.RequiresConsent(application, grant, validScopes) {
		return fiber.NewError(fiber.StatusPreconditionRequired, "consent_required")
	}

	if impersonation == nil && !application.Internal {
		if err := saveApplicationGrant(ctx, star, application, user, grant, validScopes); err != nil {
			return err
		}
	}

	var codeChallenge *string = nil
	var codeChallengeMethod *string = nil

	if challenge != "" {
		s256 := oauth.CodeChallengeMethodS256
		codeChallenge = &challenge
		codeChallengeMethod = &s256
	}

	var impersonatorID *string = nil

	if impersonation != nil {
		claimed, err := star.Database.ClaimImpersonation(ctx, impersonation.ID)
		if err != nil {
			return err
		} else if !claimed {
			return fiber.NewError(409, "This impersonation has already been used")
		}

		impersonatorID = &user.ID
		authenticatedAt = time.Now()
	}

	authCode := database.AuthorizationCode{
		ID:                  database.CreateID(database.AuthCodeKey),
		UserID:              subject.ID,
		ImpersonatorID:      impersonatorID,
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

	if impersonation != nil {
		audit.Log(ctx, audit.UserImpersonateSignIn, onApplication(application), audit.OnUser(subject), audit.With("scopes", validScopes))
	} else {
		audit.Log(ctx, audit.GrantAuthorize, onApplication(application), audit.With("scopes", validScopes))
	}

	return ctx.JSON(types.CreateApplicationAuthorizationResponse{
		Code:           authCode.Code,
		State:          authCode.State,
		Scopes:         validScopes,
		ScopedModified: len(resolution.Invalid) > 0,
	})
}
