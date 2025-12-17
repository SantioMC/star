package model

import (
	"math/rand"
)

const (
	ServiceUserID    = "service"
	ServiceUserKeyID = "key1"
)

type User struct {
	BaseModel
	Username      string    `json:"username" gorm:"uniqueIndex"`
	FirstName     *string   `json:"first_name" gorm:"uniqueIndex"`
	LastName      *string   `json:"last_name" gorm:"uniqueIndex"`
	Email         string    `json:"email" gorm:"uniqueIndex"`
	EmailVerified bool      `json:"email_verified"`
	EmailCode     string    `json:"-"`
	Roles         []Role    `json:"-" gorm:"many2many:user_roles"`
	Accounts      []Account `json:"-"`
}

var emailAlphabet = []byte("abcdefghjklnoprstvwxyz")

func createEmailCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = emailAlphabet[rand.Intn(len(emailAlphabet))]
	}
	return string(b)
}

func NewUser(username string, email string) User {
	return User{
		BaseModel: BaseModel{
			ID: CreateTypedId("usr", 12),
		},
		Username:      username,
		FirstName:     nil,
		LastName:      nil,
		Email:         email,
		EmailVerified: false,
		EmailCode:     createEmailCode(),
	}
}
