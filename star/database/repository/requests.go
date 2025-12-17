package repository

import (
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
)

type RequestStore struct {
	DB *gorm.DB
}

func (store *RequestStore) GetAuthRequestByID(id string) (*model.AuthRequest, error) {
	request := model.AuthRequest{}

	if err := store.DB.Where("id = ?", id).First(&request).Error; err != nil {
		return nil, err
	}

	return &request, nil
}

func (store *RequestStore) GetAuthRequestByCode(code string) (*model.AuthRequest, error) {
	request := model.AuthRequest{}

	if err := store.DB.Where("code = ?", code).First(&request).Error; err != nil {
		return nil, err
	}

	return &request, nil
}

func (store *RequestStore) CreateAuthRequest(request *model.AuthRequest) error {
	return store.DB.Create(request).Error
}