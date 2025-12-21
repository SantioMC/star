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

func (store *RoleStore) CreateRole(name string) (*model.Role, error) {
	role := model.NewRole(name)

	if err := store.DB.Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (store *RoleStore) GetDefaultRoles() ([]model.Role, error) {
	roles, err := gorm.G[model.Role](store.DB).Where("default = ?", true).Find(context.Background())
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (store *RoleStore) CreateSuperUserRole() (*model.Role, error) {
	role := model.NewRole("superuser")
	role.SuperRole = true

	if err := store.DB.Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (store *RoleStore) CreateDefaultRole() (*model.Role, error) {
	role := model.NewRole("default")
	role.Default = true

	if err := store.DB.Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}