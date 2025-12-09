package repository

import (
	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/crypto"
	"github.com/tubblelive/star/server/database/model"
	"gorm.io/gorm"
)

type KeyStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func decryptKeypair(keypair *model.Keypair, key [32]byte) (*model.Keypair, error) {
	privateKey, err := crypto.DecryptAES(keypair.PrivateKey, key)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.DecryptAES(keypair.PublicKey, key)
	if err != nil {
		return nil, err
	}


	keypair.PrivateKey = privateKey
	keypair.PublicKey = publicKey

	return keypair, nil
}

func (store *KeyStore) GetKeyByID(id string) (*model.Keypair, error) {
	keypair := model.Keypair{}

	if err := store.DB.Where("id = ?", id).First(&keypair).Error; err != nil {
		return nil, err
	}

	return decryptKeypair(&keypair, store.Config.AccessToken)
}

func (store *KeyStore) GetSigningKey() (*model.Keypair, error) {
	keypair := model.Keypair{}

	if err := store.DB.Where("kind = ?", model.KeypairKindSigning).First(&keypair).Error; err != nil {
		return nil, err
	}

	return decryptKeypair(&keypair, store.Config.AccessToken)
}

func (store *KeyStore) CreateKeypair(name string, kind model.KeypairKind) (*model.Keypair, error) {
	keypair, err := model.NewKeypair(name, kind)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.EncryptAES(keypair.PrivateKey, store.Config.AccessToken)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.EncryptAES(keypair.PublicKey, store.Config.AccessToken)
	if err != nil {
		return nil, err
	}

	keypair.PrivateKey = privateKey
	keypair.PublicKey = publicKey

	if err := store.DB.Create(&keypair).Error; err != nil {
		return nil, err
	}

	return decryptKeypair(keypair, store.Config.AccessToken)
}