package repository

import (
	"context"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *AppStore) GetAppByID(id string) (*model.Application, error) {
	app := model.Application{}

	if err := store.DB.Where("id = ?", id).Preload(clause.Associations).First(&app).Error; err != nil {
		return nil, err
	}

	return &app, nil
}

func (store *AppStore) GetAllApps() ([]model.Application, error) {
	return gorm.G[model.Application](store.DB).Preload("Providers", nil).Find(context.Background())
}

func (store *AppStore) GetAppBySlug(slug string) (*model.Application, error) {
	app := model.Application{}

	if err := store.DB.Where("slug = ?", slug).Preload(clause.Associations).First(&app).Error; err != nil {
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