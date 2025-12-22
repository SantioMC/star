package repository

import (
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
	app := model.Application{}

	if err := store.DB.Where("slug = ?", slug).First(&app).Error; err != nil {
		return nil, err
	}

	return &app, nil
}

func (store *AppStore) SaveApp(app *model.Application) error {
	return store.DB.Save(app).Error
}

func (store *AppStore) SaveProvider(provider *model.Provider) error {
	return store.DB.Save(provider).Error
}