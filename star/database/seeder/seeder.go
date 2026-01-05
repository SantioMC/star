package seeder

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state/scopes"
)

func SeedDatabase(firstStart bool, config *config.Config) error {
	if firstStart {
		_, err := service.CreateKeypair("signing", model.KeypairKindSigning)
		if err != nil {
			return err
		}

		_, err = service.CreateDefaultRole()
		if err != nil {
			return err
		}

		superuser, err := service.CreateSuperUserRole()
		if err != nil {
			return err
		}

		initPassword := hex.EncodeToString(crypto.CreateRand(16))
		slog.Info("creating default superuser with", "password", initPassword)
		service.CreateSuperUser(initPassword, superuser)

		app, err := service.CreateStarApp(config.Url)
		if err != nil {
			return err
		}

		for _, scope := range scopes.DefaultScopes {
			service.CreatePermission(app, scope.Node, scope.Description)
		}
	}

	// Auto register providers in our environment
	for id := range oidc.Providers {
		clientId := os.Getenv(strings.ToUpper(id) + "_CLIENT_ID")
		clientSecret := os.Getenv(strings.ToUpper(id) + "_CLIENT_SECRET")

		if clientId == "" || clientSecret == "" {
			continue
		}

		exists := service.DoesProviderExist(id)
		if exists {
			continue
		}

		slog.Info("discovered a social provider from environment", "id", id, "clientId", clientId)
		
		_, err := service.CreateProviderFromPreset(id, clientId, clientSecret)
		if err != nil {
			slog.Error("failed to create social provider", "id", id, "err", err)
			break
		}
	}

	// Update our internal application
	if !firstStart {
		application, err := service.GetAppBySlug("star")
		if err != nil {
			return err
		}

		if application == nil {
			return fmt.Errorf("failed to find internal star application, how did this happen?")
		}

		if config.Url != "" {
			url := strings.TrimSuffix(config.Url, "/")
			if !strings.Contains(url, "://") {
				url = "https://" + url
			}

			application.Url = url
			application.Callback = url + "/callback"
		}

		if err := service.UpdateApp(application); err != nil {
			return err
		}
	}

	return nil
}