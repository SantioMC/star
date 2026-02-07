package repository

import (
	"context"
	"errors"

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &app, nil
}

func (store *AppStore) GetProviderByID(id string) (*model.OAuthProvider, error) {
	provider, err := gorm.G[model.OAuthProvider](store.DB).Where("id = ?", id).First(context.Background())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &provider, err
}

func (store *AppStore) GetProviderByClientID(id string) (*model.OAuthProvider, error) {
	provider, err := gorm.G[model.OAuthProvider](store.DB).Where("client_id = ?", id).First(context.Background())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &provider, err
}

func (store *AppStore) GetAllApps() ([]model.Application, error) {
	var apps []model.Application

	if err := store.DB.Preload(clause.Associations).Find(&apps).Error; err != nil {
		return nil, err
	}

	return apps, nil
}

func (store *AppStore) GetAppBySlug(slug string) (*model.Application, error) {
	app := model.Application{}

	if err := store.DB.Where("slug = ?", slug).Preload(clause.Associations).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &app, nil
}

func (store *AppStore) GetPermissions(appId string) ([]model.Permission, error) {
	return gorm.G[model.Permission](store.DB).Where("application_id = ?", appId).Find(context.Background())
}

func (store *RoleStore) SetPermissionDescription(id string, description string) error {
	_, err := gorm.G[model.Permission](store.DB).Where("id = ?", id).Update(context.Background(), "description", description)
	return err
}

func (store *AppStore) DeleteApp(appId string) error {
	_, err := gorm.G[model.Application](store.DB).Where("id = ?", appId).Delete(context.Background())
	return err
}

func (store *AppStore) DeleteProvider(id string) error {
	_, err := gorm.G[model.OAuthProvider](store.DB).Where("id = ?", id).Delete(context.Background())
	return err
}

func (store *AppStore) SaveApp(app *model.Application) error {
	return store.DB.Save(app).Error
}

func (store *AppStore) SaveProvider(provider *model.OAuthProvider) error {
	return store.DB.Save(provider).Error
}
