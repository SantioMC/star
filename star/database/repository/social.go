package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrProviderNameInUse = fmt.Errorf("a provider already exists with the provided name!")

type SocialStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *SocialStore) GetProviderByID(id string) (*model.SocialProvider, error) {
	provider := model.SocialProvider{}

	if err := store.DB.Where("id = ?", id).First(&provider).Error; err != nil {
		return nil, err
	}

	return &provider, nil
}

func (store *SocialStore) GetProviderByName(name string) (*model.SocialProvider, error) {
	provider := model.SocialProvider{}

	if err := store.DB.Where("name = ?", name).First(&provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &provider, nil
}

func (store *SocialStore) GetProviders() ([]model.SocialProvider, error) {
	return gorm.G[model.SocialProvider](store.DB).Find(context.Background())
}

func (store *SocialStore) SaveProvider(provider *model.SocialProvider) error {
	return store.DB.Save(provider).Error
}

func (store *SocialStore) CreateState(app *model.Application, userAgent string, address string) (string, error) {
	state := crypto.CreateRandString(16)
	model := model.LoginState{
		State: state,
		Application: app,
		UserAgent: userAgent,
		Address: address,
	}

	if err := store.DB.Save(&model).Error; err != nil {
		return "", err
	}

	return state, nil
}

func (store *SocialStore) DeleteOldStates() error {
	deleteBefore := time.Now().Add(time.Hour * -1)
	_, err := gorm.G[model.LoginState](store.DB.Unscoped()).Where("created_at <= ?", deleteBefore.Unix()).Delete(context.Background())
	return err
}

func (store *SocialStore) UseState(state string) (*model.LoginState, error) {
	loginState := model.LoginState{}
	if err := store.DB.Where("state = ?", state).Preload(clause.Associations).First(&loginState).Error; err != nil {
		return nil, err
	}

	store.DB.Unscoped().Delete(&loginState)
	return &loginState, nil
}