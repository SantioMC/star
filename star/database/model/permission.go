package model

type Permission struct {
	BaseModel

	Application   *Application `json:"-"`
	ApplicationID string       `json:"role_id"`

	Permission  string `json:"scope"`
	Description string `json:"description"`
}

func NewPermission(app *Application, permission string, description string) Permission {
	return Permission{
		BaseModel: BaseModel{
			ID: CreateTypedId("per", 12),
		},

		Application: app,
		ApplicationID: app.ID,

		Permission:  permission,
		Description: description,
	}
}
