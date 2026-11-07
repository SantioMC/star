package transformer

import "github.com/tubblelive/star/internal/database"

type LoginProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Connection struct {
	LoginProvider
	Account *UserAccount `json:"account"`
}

type AdminProvider struct {
	database.Provider
	HasClientSecret bool `json:"has_client_secret"`
}

func TransformLoginProviders(providers []database.Provider) []LoginProvider {
	serialized := []LoginProvider{}

	for _, provider := range providers {
		if provider.Enabled {
			serialized = append(serialized, TransformLoginProvider(provider))
		}
	}

	return serialized
}

func TransformLoginProvider(provider database.Provider) LoginProvider {
	return LoginProvider{
		ID:   provider.ID,
		Name: provider.Name,
		Icon: provider.Icon,
	}
}

func TransformConnections(providers []database.Provider, accounts []database.Account, includeEmail bool) []Connection {
	serialized := []Connection{}

	for _, provider := range providers {
		connection := Connection{LoginProvider: TransformLoginProvider(provider)}

		for _, account := range accounts {
			if account.ProviderID != nil && *account.ProviderID == provider.ID {
				connection.Account = new(TransformAccount(account, includeEmail))
			}
		}

		if provider.Enabled || connection.Account != nil {
			serialized = append(serialized, connection)
		}
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
