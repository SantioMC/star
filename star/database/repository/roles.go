package repository

import (
	"context"

	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (store *RoleStore) GetAllRoles() ([]model.Role, error) {
	var roles []model.Role
	
	if err := store.DB.Preload(clause.Associations).Find(&roles).Error; err != nil {
		return []model.Role{}, err
	}

	return roles, nil
}

func (store *RoleStore) SetPermissions(role string, permissions []string) error {
	_, err := gorm.G[model.Role](store.DB).Where("id = ?", role).Update(context.Background(), "permissions", permissions)
	return err
}

func (store *RoleStore) RenameRole(role string, name string) error {
	_, err := gorm.G[model.Role](store.DB).Where("id = ?", role).Update(context.Background(), "name", name)
	return err
}

func (store *RoleStore) DeleteRole(role string) error {
	_, err := gorm.G[model.Role](store.DB).Where("id = ?", role).Delete(context.Background())
	return err
}

func (store *RoleStore) GetDefaultRoles() ([]model.Role, error) {
	return gorm.G[model.Role](store.DB).Where("default = ?", true).Find(context.Background())
}