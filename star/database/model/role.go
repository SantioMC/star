package model

type Role struct {
	BaseModel

	Name      string `json:"name"`
	SuperRole bool   `json:"super"`
	Global    bool   `json:"global"`

	Permission []Permission ` gorm:"many2many:role_permissions" json:"scopes"`
}

func NewRole(name string) Role {
	return Role{
		BaseModel: BaseModel{
			ID: CreateTypedId("rol", 12),
		},
		Name: name,
	}
}
