package database

import (
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/database/plugin"
	"github.com/tubblelive/star/star/database/repository"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB

	UserStore *repository.UserStore
	RoleStore *repository.RoleStore

	KeyStore     *repository.KeyStore
	AppStore     *repository.AppStore
	SocialStore  *repository.SocialStore
	RequestStore *repository.RequestStore
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
	db.AutoMigrate(&model.SocialProvider{})
	db.AutoMigrate(&model.Application{})
	db.AutoMigrate(&model.AuthRequest{})
	db.AutoMigrate(&model.Keypair{})
	db.AutoMigrate(&model.Provider{})
	db.AutoMigrate(&model.LoginState{})

	database := &Database{
		DB: db,

		UserStore: &repository.UserStore{DB: db, Config: config},
		RoleStore: &repository.RoleStore{DB: db},

		KeyStore:     &repository.KeyStore{DB: db, Config: config},
		AppStore:     &repository.AppStore{DB: db, Config: config},
		SocialStore:  &repository.SocialStore{DB: db, Config: config},
		RequestStore: &repository.RequestStore{DB: db},
	}

	if err := seedDatabase(firstStart, database); err != nil {
		return nil, err
	}

	return database, nil
}
