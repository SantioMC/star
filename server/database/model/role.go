package model

type Role struct {
	BaseModel
	Name      string
	SuperRole bool
}

func NewRole(name string) Role {
	return Role{
		BaseModel: BaseModel{
			ID: CreateTypedId("rol", 12),
		},
		Name: name,
	}
}