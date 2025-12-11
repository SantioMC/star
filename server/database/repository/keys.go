package repository

import (
	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/database/model"
	"gorm.io/gorm"
)

type KeyStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *KeyStore) GetKeyByID(id string) (*model.Keypair, error) {
	keypair := model.Keypair{}

	if err := store.DB.Where("id = ?", id).First(&keypair).Error; err != nil {
		return nil, err
	}

	return &keypair, nil
}

func (store *KeyStore) GetSigningKey() (*model.Keypair, error) {
	keypair := model.Keypair{}

	if err := store.DB.Where("kind = ?", model.KeypairKindSigning).First(&keypair).Error; err != nil {
		return nil, err
	}

	return &keypair, nil
}

func (store *KeyStore) CreateKeypair(name string, kind model.KeypairKind) (*model.Keypair, error) {
	keypair, err := model.NewKeypair(name, kind)

	if err != nil {
		return nil, err
	}

	if err := store.DB.Create(&keypair).Error; err != nil {
		return nil, err
	}

	return keypair, nil
}