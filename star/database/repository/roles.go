package repository

import (
	"context"

	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
)

type RoleStore struct {
	DB     *gorm.DB
}

func (store *RoleStore) GetRoleByID(id string) (*model.Role, error) {
	role := model.Role{}

	if err := store.DB.Where("id = ?", id).First(id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (store *RoleStore) SaveRole(role *model.Role) error {
	return store.DB.Save(role).Error
}

func (store *RoleStore) GetDefaultRoles() ([]model.Role, error) {
	roles, err := gorm.G[model.Role](store.DB).Where("default = ?", true).Find(context.Background())
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (store *RoleStore) CreatePermission(app *model.Application, permission string) error {
	perm := model.NewPermission(app, permission)
	return gorm.G[model.Permission](store.DB).Create(context.Background(), &perm)
}