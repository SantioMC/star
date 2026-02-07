package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OAuthStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *OAuthStore) GetAuthRequestByCode(code string) (*model.AuthRequest, error) {
	request := model.AuthRequest{}

	if err := store.DB.Where("code = ?", code).Preload(clause.Associations).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &request, nil
}

func (store *OAuthStore) CreateAuthRequest(
	app *model.Application,
	user *model.User,
	scopes []string,
	callbackUri string,
	state *string,
) (model.AuthRequest, error) {
	request := model.NewAuthRequest(app, user, scopes, callbackUri, state)

	if err := store.DB.Create(&request).Error; err != nil {
		return request, err
	}

	return request, nil
}

func (store *OAuthStore) MarkCompleted(code string) error {
	_, err := gorm.G[model.AuthRequest](store.DB).
		Where("code = ?", code).
		Update(context.Background(), "completed_at", time.Now().Unix())

	return err
}

func (store *OAuthStore) DeleteOldAuthRequests() error {
	old := time.Now().Add(-time.Hour * 24)

	_, err := gorm.G[model.AuthRequest](store.DB).
		Where("created_at < ?", old).
		Delete(context.Background())

	return err
}
