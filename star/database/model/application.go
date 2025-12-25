package model

import (
	"strings"
)

type Application struct {
	BaseModel

	Name     string `json:"name" gorm:"unique"`
	Slug     string `json:"slug" gorm:"unique"`
	Url      string `json:"url"`
	Callback string `json:"callback"`
	Internal bool   `json:"internal"`

	Providers []Provider `json:"providers"`
}

func NewApplication(name string) Application {
	slug := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	return Application{
		BaseModel: BaseModel{
			ID: CreateTypedId("app", 12),
		},

		Name:     name,
		Slug:     slug,
		Url:      "",
		Callback: "",

		Internal: false,

		Providers: []Provider{},
	}
}
