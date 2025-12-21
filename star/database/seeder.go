package database

import (
	"encoding/hex"
	"log/slog"
	"os"
	"strings"

	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
	"github.com/tubblelive/star/star/database/model"
)

func seedDatabase(firstStart bool, db *Database) error {
	if firstStart {
		signingKey, err := model.NewKeypair("signing", model.KeypairKindSigning)
		if err != nil {
			return err
		}

		_, err = db.KeyStore.CreateKeypair(signingKey.Name, signingKey.Kind)
		if err != nil {
			return err
		}

		_, err = db.RoleStore.CreateDefaultRole()
		if err != nil {
			return err
		}

		superuser, err := db.RoleStore.CreateSuperUserRole()
		if err != nil {
			return err
		}

		initPassword := hex.EncodeToString(crypto.CreateRand(16))
		slog.Info("creating default superuser with", "password", initPassword)
		db.UserStore.CreateSuperUser(initPassword, superuser)

		_, err = db.AppStore.CreateStarApp()
		if err != nil {
			return err
		}
	}

	// Auto register providers in our environment
	for id := range oidc.Providers {
		clientId := os.Getenv(strings.ToUpper(id) + "_CLIENT_ID")
		clientSecret := os.Getenv(strings.ToUpper(id) + "_CLIENT_SECRET")

		if clientId == "" || clientSecret == "" {
			continue
		}

		exists := db.SocialStore.DoesProviderExist(id)
		if exists {
			continue
		}

		slog.Info("discovered a social provider from environment", "id", id, "clientId", clientId)
		
		_, err := db.SocialStore.CreateProviderFromPreset(id, clientId, clientSecret)
		if err != nil {
			slog.Error("failed to create social provider", "id", id, "err", err)
			break
		}
	}

	return nil
}