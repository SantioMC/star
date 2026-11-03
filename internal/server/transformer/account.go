package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type UserAccount struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	ExternalID *string   `json:"account_id"`
	Name       *string   `json:"name"`
	Email      *string   `json:"email"`
	Avatar     *string   `json:"picture"`
	CreatedAt  time.Time `json:"created_at"`
}

func TransformAccount(account database.Account, includeEmail bool) UserAccount {
	email := account.Email
	if !includeEmail {
		email = nil
	}

	return UserAccount{
		ID:         account.ID,
		Type:       account.Type,
		Avatar:     account.Avatar,
		Name:       account.Name,
		ExternalID: account.ExternalID,
		Email:      email,
		CreatedAt:  account.CreatedAt,
	}
}

func TransformAccounts(accounts []database.Account, includeEmail bool) []UserAccount {
	serialized := []UserAccount{}

	for _, value := range accounts {
		account := TransformAccount(value, includeEmail)
		serialized = append(serialized, account)
	}

	return serialized
}
