package database

import (
	"log/slog"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/database/plugin"
	"github.com/tubblelive/star/star/database/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB

	UserStore    *repository.UserStore
	SessionStore *repository.SessionStore
	RoleStore    *repository.RoleStore

	KeyStore     *repository.KeyStore
	AppStore     *repository.AppStore
	SocialStore  *repository.SocialStore
	RequestStore *repository.RequestStore
}

func Connect(dialector gorm.Dialector, config *config.Config) (*Database, bool, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.NewSlogLogger(slog.Default(), logger.Config{
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Warn,
		}),
	})

	if err != nil {
		return nil, false, err
	}

	db.Use(plugin.CreateEncryptionPlugin(config.AccessToken))
	firstStart := !db.Migrator().HasTable(&model.User{})

	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Account{})
	db.AutoMigrate(&model.Role{})
	db.AutoMigrate(&model.Session{})
	db.AutoMigrate(&model.SocialProvider{})
	db.AutoMigrate(&model.Application{})
	db.AutoMigrate(&model.AuthRequest{})
	db.AutoMigrate(&model.Keypair{})
	db.AutoMigrate(&model.OAuthProvider{})
	db.AutoMigrate(&model.LoginState{})
	db.AutoMigrate(&model.Suspension{})
	db.AutoMigrate(&model.Permission{})

	database := &Database{
		DB: db,

		UserStore:    &repository.UserStore{DB: db, Config: config},
		SessionStore: &repository.SessionStore{DB: db},
		RoleStore:    &repository.RoleStore{DB: db},

		KeyStore:     &repository.KeyStore{DB: db, Config: config},
		AppStore:     &repository.AppStore{DB: db, Config: config},
		SocialStore:  &repository.SocialStore{DB: db, Config: config},
		RequestStore: &repository.RequestStore{DB: db},
	}

	return database, firstStart, nil
}