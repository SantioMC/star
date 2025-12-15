package database

import (
	"encoding/hex"
	"log/slog"

	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/crypto"
	"github.com/tubblelive/star/server/database/model"
	"github.com/tubblelive/star/server/database/plugin"
	"github.com/tubblelive/star/server/database/repository"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB

	UserStore *repository.UserStore
	RoleStore *repository.RoleStore

	KeyStore     *repository.KeyStore
	AppStore     *repository.AppStore
	RequestStore *repository.RequestStore
}

func seedDatabase(db *Database) error {
	signingKey, err := model.NewKeypair("signing", model.KeypairKindSigning)
	if err != nil {
		return err
	}

	_, err = db.KeyStore.CreateKeypair(signingKey.Name, signingKey.Kind)
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

	return nil
}

func Connect(dialector gorm.Dialector, config *config.Config) (*Database, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.Use(plugin.CreateEncryptionPlugin(config.AccessToken))
	firstStart := !db.Migrator().HasTable(&model.User{})

	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Account{})
	db.AutoMigrate(&model.Role{})
	db.AutoMigrate(&model.RefreshToken{})
	db.AutoMigrate(&model.Application{})
	db.AutoMigrate(&model.AuthRequest{})
	db.AutoMigrate(&model.Keypair{})

	database := &Database{
		DB: db,

		UserStore: &repository.UserStore{DB: db, Config: config},
		RoleStore: &repository.RoleStore{DB: db},

		KeyStore:     &repository.KeyStore{DB: db, Config: config},
		AppStore:     &repository.AppStore{DB: db, Config: config},
		RequestStore: &repository.RequestStore{DB: db},
	}

	if firstStart {
		if err := seedDatabase(database); err != nil {
			return nil, err
		}
	}

	return database, nil
}
