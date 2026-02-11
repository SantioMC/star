package service

import (
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

func GetKeyByID(id string) (*model.Keypair, error) {
	return state.Database.KeyStore.GetKeyByID(id)
}

func GetSigningKey() (*model.Keypair, error) {
	return state.Database.KeyStore.GetKeyByKind(model.KeypairKindSigning)
}

func CreateKeypair(name string, kind model.KeypairKind) (*model.Keypair, error) {
	return state.Database.KeyStore.CreateKeypair(name, kind)
}
