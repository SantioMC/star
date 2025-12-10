package repository

import (
	"fmt"

	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/crypto"
	"github.com/tubblelive/star/server/database/model"
	"gorm.io/gorm"
)

type AppStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func decryptApp(app *model.Application, key [32]byte) (*model.Application, error) {
	if app.Secret == "" {
		return app, nil
	}

	secret, err := crypto.DecryptAES(app.Secret, key)
	if err != nil {
		return nil, err
	}

	app.Secret = secret
	return app, nil
}

func (store *AppStore) GetAppByID(id string) (*model.Application, error) {
	app := model.Application{}

	if err := store.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}

	return decryptApp(&app, store.Config.AccessToken)
}

func (store *AppStore) GetAppBySlug(slug string) (*model.Application, error) {
	if slug == "" {
		return nil, fmt.Errorf("provided slug was empty")
	}

	app := model.Application{}

	if err := store.DB.Where("slug = ?", slug).First(&app).Error; err != nil {
		return nil, err
	}

	return decryptApp(&app, store.Config.AccessToken)
}

func (store *AppStore) SaveApp(app *model.Application) error {
	secret, err := crypto.EncryptAES(app.Secret, store.Config.AccessToken)
	if err != nil {
		return err
	}

	app.Secret = secret
	return store.DB.Save(app).Error
}

func (store *AppStore) CreateStarApp() (*model.Application, error) {
	app := model.NewApplication("Star")
	app.Internal = true

	err := store.SaveApp(&app)
	if err != nil {
		return nil, err
	}
	
	return decryptApp(&app, store.Config.AccessToken)
}
