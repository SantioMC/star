package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/verifier"
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
	if name == "" {
		return nil, fmt.Errorf("no name was provided")
	}
	
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

func (store *SocialStore) DiscoverAndCreateProvider(
	name string, 
	discoveryUri string, 
	clientId string, 
	clientSecret string,
) (*model.SocialProvider, error) {
	if err := verifier.VerifyProviderName(name); err != nil {
		return nil, err
	}

	provider, err := model.NewSocialProviderFromDiscovery(name, discoveryUri, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	existing, err := store.GetProviderByName(name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrProviderNameInUse
	}

	if err := store.DB.Save(provider).Error; err != nil {
		return nil, err
	}

	return provider, nil
}

func (store *SocialStore) DoesProviderExist(name string) bool {
	var exists bool
	if err := store.DB.Model(&model.SocialProvider{}).Select("COUNT(*) > 0").Where("name = ?", name).First(&exists).Error; err != nil {
		return false
	}

	return exists
}

func (store *SocialStore) CreateProviderFromPreset(
	name string, 
	clientId string, 
	clientSecret string,
) (*model.SocialProvider, error) {
	if err := verifier.VerifyProviderName(name); err != nil {
		return nil, err
	}

	config, ok := oidc.Providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown preset provider %s", name)
	}

	provider, err := model.NewSocialProviderFromPreset(name, config, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	existing, err := store.GetProviderByName(name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrProviderNameInUse
	}

	if err := store.DB.Save(provider).Error; err != nil {
		return nil, err
	}

	return provider, nil
}

func (store *SocialStore) CreateState(app *model.Application) (string, error) {
	state := crypto.CreateRandString(16)
	model := model.LoginState{
		State: state,
		Application: app,
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

func (store *SocialStore) UseState(state string) (*model.Application, error) {
	loginState := model.LoginState{}
	if err := store.DB.Where("state = ?", state).Preload(clause.Associations).First(&loginState).Error; err != nil {
		return nil, err
	}

	store.DB.Unscoped().Delete(&loginState)
	return loginState.Application, nil
}