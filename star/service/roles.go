package service

import (
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
)

func GetRoleByID(id string) (*model.Role, error) {
	return state.Database.RoleStore.GetRoleByID(id)
}

func CreateRole(name string) (*model.Role, error) {
	role := model.NewRole(name)
	
	if err := state.Database.RoleStore.SaveRole(&role); err != nil {
		return nil, err
	}

	return &role, nil
}

func GetDefaultRoles() ([]model.Role, error) {
	return state.Database.RoleStore.GetDefaultRoles()
}

func GetAllRoles() ([]model.Role, error) {
	return state.Database.RoleStore.GetAllRoles()
}

func CreateSuperUserRole() (*model.Role, error) {
	role := model.NewRole("superuser")
	role.SuperRole = true

	if err := state.Database.RoleStore.SaveRole(&role); err != nil {
		return nil, err
	}
	
	return &role, nil
}

func CreateDefaultRole() (*model.Role, error) {
	role := model.NewRole("default")
	role.Global = true

	if err := state.Database.RoleStore.SaveRole(&role); err != nil {
		return nil, err
	}
	
	return &role, nil
}

func CreatePermission(role *model.Application, permission string) error {
	return state.Database.RoleStore.CreatePermission(role, permission)
}