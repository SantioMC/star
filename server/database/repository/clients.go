package repository

import (
	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/crypto"
	"github.com/tubblelive/star/server/database/model"
	"gorm.io/gorm"
)

type ClientStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func decryptClient(client *model.Application, key [32]byte) (*model.Application, error) {
	if client.Secret == "" {
		return client, nil
	}

	secret, err := crypto.DecryptAES(client.Secret, key)
	if err != nil {
		return nil, err
	}

	client.Secret = secret
	return client, nil
}

func (store *ClientStore) GetClientByID(id string) (*model.Application, error) {
	client := model.Application{}

	if err := store.DB.Where("id = ?", id).First(&client).Error; err != nil {
		return nil, err
	}

	return decryptClient(&client, store.Config.AccessToken)
}

func (store *ClientStore) SaveClient(client *model.Application) error {
	secret, err := crypto.EncryptAES(client.Secret, store.Config.AccessToken)
	if err != nil {
		return err
	}

	client.Secret = secret
	return store.DB.Save(client).Error
}

func (store *ClientStore) CreateStarClient() (*model.Application, error) {
	client := model.NewApplication("Star")
	client.Internal = true

	err := store.SaveClient(&client)
	if err != nil {
		return nil, err
	}
	
	return decryptClient(&client, store.Config.AccessToken)
}
