package service

import (
	"log/slog"

	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/util"
	"gorm.io/datatypes"
)

func GetRoleByID(id string) (*model.Role, error) {
	return state.Database.RoleStore.GetRoleByID(id)
}

func CreateRole(name string, permissions *[]string) (*model.Role, error) {
	role := model.NewRole(name)

	if permissions != nil {
		role.Permissions = datatypes.NewJSONType(util.Unique(*permissions))
	}
	
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

func SetPermissions(role string, permissions []string) error {
	return state.Database.RoleStore.SetPermissions(role, util.Unique(permissions))
}

func RenameRole(role string, name string) error {
	return state.Database.RoleStore.RenameRole(role, name)
}

func DeleteRole(role string) error {
	return state.Database.RoleStore.DeleteRole(role)
}

func CreateSuperUserRole() (*model.Role, error) {
	role := model.NewRole("superuser")
	role.Description = "The default superuser role, permitting all actions"
	role.SuperRole = true

	if err := state.Database.RoleStore.SaveRole(&role); err != nil {
		return nil, err
	}
	
	return &role, nil
}

func CreateDefaultRole() (*model.Role, error) {
	role := model.NewRole("default")
	role.Description = "The default role, given to all users automatically"
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

func FindNonDefaultRolesById(roles []string) ([]model.Role, error) {
	return state.Database.RoleStore.FindNonDefaultRolesById(roles)
}