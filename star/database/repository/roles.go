package repository

import (
	"context"
	"errors"

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

func (store *RoleStore) FindNonDefaultRolesById(ids []string) ([]model.Role, error) {
	var roles []model.Role
	
	if err := store.DB.Where("id IN ? AND global = ?", ids, false).Preload(clause.Associations).Find(&roles).Error; err != nil {
		return []model.Role{}, err
	}

	return roles, nil
}

func (store *RoleStore) SetPermissions(role string, permissions []string) error {
	_, err := gorm.G[model.Role](store.DB).Where("id = ?", role).Update(context.Background(), "permissions", permissions)
	return err
}

func (store *RoleStore) GetPermissionByID(id string) (*model.Permission, error) {
	permission, err := gorm.G[model.Permission](store.DB).Preload("Application", nil).Where("id = ?", id).First(context.Background())

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &permission, nil
}

func (store *RoleStore) CreatePermission(app *model.Application, permission string, description string) (model.Permission, error) {
	perm := model.NewPermission(app, permission, description)
	return perm, gorm.G[model.Permission](store.DB).Create(context.Background(), &perm)
}

func (store *RoleStore) DeletePermission(id string) error {
	_, err := gorm.G[model.Permission](store.DB).Where("id = ?", id).Delete(context.Background())
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