package service

import (
	"log/slog"

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

func CreatePermission(application *model.Application, permission string, description string) error {
	return state.Database.RoleStore.CreatePermission(application, permission, description)
}

func GetPermissions(appId string) []model.Permission {
	permissions, err := state.Database.AppStore.GetPermissions(appId)
	if err != nil {
		slog.Error("failed to get permissions for", "app", appId)
		return []model.Permission{}
	}

	return permissions
}