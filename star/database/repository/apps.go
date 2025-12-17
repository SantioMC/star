package repository

import (
	"fmt"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
)

type AppStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *AppStore) GetAppByID(id string) (*model.Application, error) {
	app := model.Application{}

	if err := store.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}

	return &app, nil
}

func (store *AppStore) GetAppBySlug(slug string) (*model.Application, error) {
	if slug == "" {
		return nil, fmt.Errorf("provided slug was empty")
	}

	app := model.Application{}

	if err := store.DB.Where("slug = ?", slug).First(&app).Error; err != nil {
		return nil, err
	}

	return &app, nil
}

func (store *AppStore) CreateStarApp() (*model.Application, error) {
	// Create internal Star app
	app := model.NewApplication("Star")
	app.Internal = true

	if store.Config.Url != "" {
		app.Url = store.Config.Url
	}

	err := store.DB.Save(&app).Error
	if err != nil {
		return nil, err
	}

	// Create OAuth2 provider for star app
	provider := model.NewProvider(&app, model.OAuth2)
	err = store.DB.Save(&provider).Error
	if err != nil {
		return nil, err
	}

	return &app, nil
}
