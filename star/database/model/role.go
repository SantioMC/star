package model

type Role struct {
	BaseModel
	Name      string `json:"name"`
	SuperRole bool   `json:"super"`
	Default   bool   `json:"default"`
}

func NewRole(name string) Role {
	return Role{
		BaseModel: BaseModel{
			ID: CreateTypedId("rol", 12),
		},
		Name: name,
	}
}
