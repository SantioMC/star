package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/federation"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/oauth"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func getLoginProviders(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	return ctx.JSON(transformer.TransformLoginProviders(star.Database.GetProviders(ctx)))
}

func providerCallbackURL(star *state.StarServer, provider *database.Provider) string {
	return star.Config.ServerURL + "/api/user/login/provider/" + url.PathEscape(provider.ID) + "/callback"
}

func setProviderCookie(ctx fiber.Ctx, star *state.StarServer, name string, value string, maxAge int) {
	sameSite := fiber.CookieSameSiteLaxMode
	if request.SecureCookies(star) {
		sameSite = fiber.CookieSameSiteNoneMode
	}

	request.SetCookie(ctx, star, fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/api/user/login/provider",
		MaxAge:   maxAge,
		SameSite: sameSite,
	})
}

func redirectToLogin(ctx fiber.Ctx, star *state.StarServer, key string, value string) error {
	query := url.Values{}
	query.Set(key, value)

	return ctx.Redirect().To(star.Config.ServerURL + "/login?" + query.Encode())
}

func startProviderLogin(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	provider := star.Database.GetProviderByID(ctx, ctx.Params("provider"))
	if provider == nil {
		return redirectToLogin(ctx, star, "error", "This sign in provider is no longer available")
	}

	endpoints, err := federation.ResolveEndpoints(ctx, provider)
	if err != nil {
		slog.Error("failed to resolve provider endpoints", "provider", provider.ID, "err", err)
		return redirectToLogin(ctx, star, "error", "Unable to reach the sign in provider, please try again later")
	}

	loginState := types.ProviderLoginState{
		Provider:  provider.ID,
		State:     cryptography.CreateSecureString(32),
		Verifier:  cryptography.CreateSecureString(64),
		Nonce:     cryptography.CreateSecureString(32),
		CreatedAt: time.Now(),
	}

	encoded, err := json.Marshal(loginState)
	if err != nil {
		return err
	}

	sealed, err := cryptography.EncryptAES(string(encoded), star.Config.AppToken)
	if err != nil {
		return err
	}

	target, err := federation.AuthorizationURL(endpoints, provider, &federation.Authorization{
		RedirectURI: providerCallbackURL(star, provider),
		State:       loginState.State,
		Nonce:       loginState.Nonce,
		Challenge:   oauth.CreateCodeChallenge(loginState.Verifier),
	})

	if err != nil {
		slog.Error("failed to build provider authorization url", "provider", provider.ID, "err", err)
		return redirectToLogin(ctx, star, "error", "This sign in provider is misconfigured, please contact an administrator")
	}

	setProviderCookie(ctx, star, "star_provider_login", sealed, 600)
	return ctx.Redirect().To(target)
}

func readProviderLoginState(ctx fiber.Ctx, star *state.StarServer) (*types.ProviderLoginState, error) {
	sealed := ctx.Cookies("star_provider_login")
	if sealed == "" {
		return nil, errors.New("missing provider login cookie")
	}

	decrypted, err := cryptography.DecryptAES(sealed, star.Config.AppToken)
	if err != nil {
		return nil, err
	}

	var loginState types.ProviderLoginState
	if err := json.Unmarshal([]byte(decrypted), &loginState); err != nil {
		return nil, err
	}

	return &loginState, nil
}

func callbackValue(ctx fiber.Ctx, key string) string {
	if ctx.Method() == fiber.MethodPost {
		return ctx.FormValue(key)
	}

	return ctx.Query(key)
}

func providerLoginCallback(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	loginState, err := readProviderLoginState(ctx, star)
	setProviderCookie(ctx, star, "star_provider_login", "", -1)

	if err != nil || loginState.Provider != ctx.Params("provider") || time.Since(loginState.CreatedAt) > time.Minute*10 {
		return redirectToLogin(ctx, star, "error", "Your sign in attempt expired, please try again")
	}

	if subtle.ConstantTimeCompare([]byte(callbackValue(ctx, "state")), []byte(loginState.State)) != 1 {
		return redirectToLogin(ctx, star, "error", "Your sign in attempt expired, please try again")
	}

	if reason := callbackValue(ctx, "error"); reason == "access_denied" {
		return redirectToLogin(ctx, star, "error", "Sign in was cancelled")
	} else if reason != "" {
		slog.Warn("provider returned an authorization error", "provider", loginState.Provider, "error", reason)
		return redirectToLogin(ctx, star, "error", "The sign in provider rejected the request, please try again")
	}

	user, account, err := authenticateWithProvider(ctx, star, loginState, callbackValue(ctx, "code"))
	if err != nil {
		if known, ok := errors.AsType[*fiber.Error](err); ok {
			return redirectToLogin(ctx, star, "error", known.Message)
		}

		slog.Error("provider login failed", "provider", loginState.Provider, "err", err)
		return redirectToLogin(ctx, star, "error", "Unable to sign you in with this provider, please try again later")
	}

	token := database.CreateLoginFlowToken()
	if err := star.Database.CreateLoginFlow(ctx, &database.LoginFlow{
		ID:        database.CreateID(database.LoginFlowKey),
		UserID:    user.ID,
		AccountID: &account.ID,
		Token:     token,
		Stage:     database.StageProvider,
		CreatedAt: time.Now(),
	}); err != nil {
		return err
	}

	setProviderCookie(ctx, star, "star_provider_flow", token, 300)
	return redirectToLogin(ctx, star, "flow", token)
}

func authenticateWithProvider(
	ctx fiber.Ctx,
	star *state.StarServer,
	loginState *types.ProviderLoginState,
	code string,
) (*database.User, *database.Account, error) {
	if code == "" {
		return nil, nil, errors.New("provider did not return an authorization code")
	}

	provider := star.Database.GetProviderByID(ctx, loginState.Provider)
	if provider == nil {
		return nil, nil, fiber.NewError(404, "This sign in provider is no longer available")
	}

	endpoints, err := federation.ResolveEndpoints(ctx, provider)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := federation.Exchange(ctx, endpoints, provider, code, providerCallbackURL(star, provider), loginState.Verifier)
	if err != nil {
		return nil, nil, err
	}

	identity, err := federation.FetchIdentity(ctx, endpoints, provider, tokens, loginState.Nonce)
	if err != nil {
		return nil, nil, err
	}

	return federation.ResolveAccount(ctx, star, provider, identity)
}

func completeProviderLogin(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.CompleteProviderLoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	expired := fiber.NewError(400, "Your sign in attempt expired, please try again")

	binding := ctx.Cookies("star_provider_flow")
	setProviderCookie(ctx, star, "star_provider_flow", "", -1)

	if subtle.ConstantTimeCompare([]byte(binding), []byte(request.Flow)) != 1 {
		return expired
	}

	loginFlow := star.Database.GetLoginStageByToken(ctx, request.Flow)
	if loginFlow == nil || loginFlow.Stage != database.StageProvider || loginFlow.AccountID == nil {
		return expired
	}

	claimed, err := star.Database.DeleteLoginFlowByID(ctx, loginFlow.ID)
	if err != nil {
		return err
	} else if !claimed || time.Since(loginFlow.CreatedAt) > time.Minute*5 {
		return expired
	}

	user := star.Database.GetUserByID(ctx, loginFlow.UserID)
	account := star.Database.GetAccountByID(ctx, *loginFlow.AccountID)

	if user == nil || account == nil {
		return expired
	}

	return continueLogin(ctx, star, user, account)
}
