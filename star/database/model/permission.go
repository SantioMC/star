package model

type Permission struct {
	BaseModel

	Application   *Application `json:"-"`
	ApplicationID string       `json:"app_id"`

	Permission string `json:"scope"`
}

func NewPermission(app *Application, permission string) Permission {
	return Permission{
		BaseModel: BaseModel{
			ID: CreateTypedId("per", 12),
		},

		Application: app,
		Permission:  permission,
	}
}
