package transformer

import "github.com/tubblelive/star/internal/database"

type LoginProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type AdminProvider struct {
	database.Provider
	HasClientSecret bool `json:"has_client_secret"`
}

func TransformLoginProviders(providers []database.Provider) []LoginProvider {
	serialized := []LoginProvider{}

	for _, provider := range providers {
		serialized = append(serialized, LoginProvider{
			ID:   provider.ID,
			Name: provider.Name,
			Icon: provider.Icon,
		})
	}

	return serialized
}

func TransformProvider(provider database.Provider) AdminProvider {
	return AdminProvider{
		Provider:        provider,
		HasClientSecret: provider.ClientSecret != "",
	}
}

func TransformProviders(providers []database.Provider) []AdminProvider {
	transformed := make([]AdminProvider, len(providers))
	for idx, provider := range providers {
		transformed[idx] = TransformProvider(provider)
	}

	return transformed
}
