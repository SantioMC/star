package repository

import (
	"fmt"
	"strings"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
)

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
		return nil, err
	}

	return &provider, nil
}

func (store *SocialStore) DiscoverAndCreateProvider(
	name string, 
	discoveryUri string, 
	clientId string, 
	clientSecret string,
) (*model.SocialProvider, error) {
	provider, err := model.NewSocialProviderFromDiscovery(name, discoveryUri, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	if err := store.DB.Save(provider).Error; err != nil {
		return nil, err
	}

	return provider, nil
}

func (store *SocialStore) CreateProviderFromPreset(
	name string, 
	clientId string, 
	clientSecret string,
) (*model.SocialProvider, error) {
	config, ok := oidc.Providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown preset provider %s", name)
	}

	provider, err := model.NewSocialProviderFromPreset(name, config, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	if err := store.DB.Save(provider).Error; err != nil {
		return nil, err
	}

	return provider, nil
}
