package model

import "gorm.io/datatypes"

type Role struct {
	BaseModel

	Name        string `json:"name"`
	Description string `json:"description"`

	SuperRole bool `json:"super"`
	Global    bool `json:"global"`

	Permissions datatypes.JSONType[[]string] `json:"permissions"`
}

func NewRole(name string) Role {
	return Role{
		BaseModel: BaseModel{
			ID: CreateTypedId("rol", 12),
		},

		Name:        name,
		Description: "",

		Permissions: datatypes.NewJSONType([]string{}),
	}
}
