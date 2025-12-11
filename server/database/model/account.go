package model

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/tubblelive/star/server/crypto"
)

type Account struct {
	BaseModel
	UserID   string
	User     *User
	Username string
	Provider string
	Password *string `json:"-" encrypt:"true"`
	Token    *string `json:"-" encrypt:"true"`
}

func NewAccountWithUserPass(user *User, username string, password string) Account {
	hash := crypto.HashString(password)

	return Account{
		BaseModel: BaseModel{
			ID: CreateTypedId("acc", 12),
		},
		Provider: "internal",
		User:     user,
		Username: username,
		Password: &hash,
		Token:    nil,
	}
}

func NewAccountWithToken(user *User, provider string, token string) (Account, error) {
	if provider == "internal" {
		slog.Error("attempted to create internal account with token")
		return Account{}, fmt.Errorf("cannot create internal account with token")
	}

	hash := crypto.HashString(token)

	return Account{
		BaseModel: BaseModel{
			ID: CreateTypedId("acc", 12),
		},
		Provider: strings.ToLower(provider),
		User:     user,
		Username: user.Username,
		Password: nil,
		Token:    &hash,
	}, nil
}
