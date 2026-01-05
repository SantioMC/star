package service

import (
	"fmt"
	"strings"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/verifier"
)

var ErrProviderNameInUse = fmt.Errorf("a provider already exists with the provided name!")

func GetProviderByID(id string) (*model.SocialProvider, error) {
	return state.Database.SocialStore.GetProviderByID(id)
}

func GetProviderByName(name string) (*model.SocialProvider, error) {
	if name == "" {
		return nil, fmt.Errorf("no name was provided")
	}
	
	return state.Database.SocialStore.GetProviderByName(name)
}

func GetProviders() ([]model.SocialProvider, error) {
	return state.Database.SocialStore.GetProviders()
}

func DiscoverAndCreateProvider(
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

	existing, err := GetProviderByName(name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrProviderNameInUse
	}

	if err := state.Database.SocialStore.SaveProvider(provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func DoesProviderExist(name string) bool {
	provider, err := state.Database.SocialStore.GetProviderByName(name)
	if err != nil {
		return false
	}

	return provider != nil
}

func CreateProviderFromPreset(
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

	existing, err := GetProviderByName(name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrProviderNameInUse
	}

	if err := state.Database.SocialStore.SaveProvider(provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func CreateState(app *model.Application, userAgent string, address string) (string, error) {
	return state.Database.SocialStore.CreateState(app, userAgent, address)
}

func DeleteOldStates() error {
	return state.Database.SocialStore.DeleteOldStates()
}

func UseState(code string) (*model.LoginState, error) {
	return state.Database.SocialStore.UseState(code)
}