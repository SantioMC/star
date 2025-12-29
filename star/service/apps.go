package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
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

func GetAppProviderByID(id string) (*model.Provider, error) {
	provider, err := state.Database.AppStore.GetProviderByID(id)
	if err != nil {
		return &model.Provider{}, err
	}
	
	return provider, nil
}

func CreateAppProvider(app *model.Application, kind model.ProviderKind) (model.Provider, error) {
	if (!slices.Contains(model.AllProviderKinds, kind)) {
		return model.Provider{}, fmt.Errorf("unknown provider kind")
	}
	
	provider := model.NewProvider(app, kind)
	if err := state.Database.AppStore.SaveProvider(&provider); err != nil {
		return model.Provider{}, err
	}

	return provider, nil
}

func DeleteAppProvider(id string) error {
	return state.Database.AppStore.DeleteProvider(id)
}

func RotateAppProvider(provider model.Provider) error {
	provider.RotateSecret()
	return state.Database.AppStore.SaveProvider(&provider)
}