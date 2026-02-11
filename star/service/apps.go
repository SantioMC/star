package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
	"gorm.io/datatypes"
)

func GetAppByID(id string) (*model.Application, error) {
	return state.Database.AppStore.GetAppByID(id)
}

func GetAllApps() ([]model.Application, error) {
	return state.Database.AppStore.GetAllApps()
}

func GetAppBySlug(slug string) (*model.Application, error) {
	if slug == "" {
		return nil, fmt.Errorf("provided slug was empty")
	}

	return state.Database.AppStore.GetAppBySlug(slug)
}

func CreateApplication(name string, slug *string, url string) (*model.Application, error) {
	// Create internal Star app
	app := model.NewApplication(name)
	app.Url = url

	sanitized := strings.TrimSpace(strings.ToLower(*slug))
	if slug != nil && sanitized != "" {
		app.Slug = *slug
	}

	err := state.Database.AppStore.SaveApp(&app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func CreateStarApp(url string) (*model.Application, error) {
	// Create internal Star app
	app := model.NewApplication("Star")
	app.Internal = true

	if url != "" {
		app.Url = url
	}

	// Set to `/callback` if no url is specified
	app.Callback = strings.TrimSuffix(app.Url, "/") + "/callback"

	// Set up our oidc provider
	provider := model.NewOAuthProvider(&app)
	provider.ApplicationType = model.AppTypePublic
	provider.Fingerprinting = true

	app.OauthProvider = &provider
	app.OauthProviderID = &provider.ID

	err := state.Database.AppStore.SaveApp(&app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func DeleteApp(appId string) error {
	return state.Database.AppStore.DeleteApp(appId)
}

func UpdateApp(app *model.Application) error {
	return state.Database.AppStore.SaveApp(app)
}

func GetAppProviderByID(id string) (*model.OAuthProvider, error) {
	provider, err := state.Database.AppStore.GetProviderByID(id)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func GetAppProviderByClientID(id string) (*model.OAuthProvider, error) {
	provider, err := state.Database.AppStore.GetProviderByClientID(id)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func CreateAppOAuthProvider(app *model.Application) (model.OAuthProvider, error) {
	provider := model.NewOAuthProvider(app)
	app.OauthProvider = &provider
	app.OauthProviderID = &provider.ID

	if err := state.Database.AppStore.SaveApp(app); err != nil {
		return provider, err
	}

	return provider, nil
}

func DeleteAppProvider(id string) error {
	return state.Database.AppStore.DeleteProvider(id)
}

func RotateAppProvider(provider model.OAuthProvider) error {
	provider.RotateSecret()
	return state.Database.AppStore.SaveProvider(&provider)
}

func CreateRedirrect(provider model.OAuthProvider, redirectUri string) error {
	redirects := provider.RedirectURIs.Data()
	sanitized := strings.TrimSpace(strings.ToLower(redirectUri))

	if slices.Contains(redirects, sanitized) {
		return fmt.Errorf("redirect uri already exists")
	}

	redirects = append(redirects, sanitized)
	provider.RedirectURIs = datatypes.NewJSONType(redirects)

	return state.Database.AppStore.SaveProvider(&provider)
}

func DeleteRedirect(provider model.OAuthProvider, redirectUri string) error {
	redirects := provider.RedirectURIs.Data()
	sanitized := strings.TrimSpace(strings.ToLower(redirectUri))

	if !slices.Contains(redirects, sanitized) {
		return fmt.Errorf("redirect uri does not exist")
	}

	for i, uri := range redirects {
		if uri == sanitized {
			redirects = append(redirects[:i], redirects[i+1:]...)
			break
		}
	}

	provider.RedirectURIs = datatypes.NewJSONType(redirects)
	return state.Database.AppStore.SaveProvider(&provider)
}

func SaveProvider(provider model.OAuthProvider) error {
	return state.Database.AppStore.SaveProvider(&provider)
}
