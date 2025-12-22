package service

import (
	"fmt"
	"strings"

	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

func GetAppByID(id string) (*model.Application, error) {
	return state.Database.AppStore.GetAppByID(id)
}

func GetAppBySlug(slug string) (*model.Application, error) {
	if slug == "" {
		return nil, fmt.Errorf("provided slug was empty")
	}
	
	return state.Database.AppStore.GetAppBySlug(slug)
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

	// Create OAuth2 provider for star app
	provider := model.NewProvider(&app, model.OAuth2)
	if err := state.Database.AppStore.SaveProvider(&provider); err != nil {
		return nil, err
	}

	return &app, nil
}
