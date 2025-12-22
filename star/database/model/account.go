package model

import (
	"github.com/tubblelive/star/oidc"
	"github.com/tubblelive/star/star/crypto"
)

type Account struct {
	BaseModel

	UserID string `json:"user_id"`
	User   *User  `json:"-"`

	ProviderID *string         `json:"provider_id"`
	Provider   *SocialProvider `json:"-"`

	Username string  `json:"username"`
	Password *string `json:"-"`

	PlatformID     string `json:"platform_id"`
	PlatformAvatar string `json:"platform_avatar"`
}

func NewInternalAccount(user *User, username string, password string) Account {
	hash := crypto.HashString(password)

	return Account{
		BaseModel: BaseModel{
			ID: CreateTypedId("acc", 12),
		},

		Provider:   nil,
		ProviderID: nil,

		User:   user,
		UserID: user.ID,

		Username: username,
		Password: &hash,
	}
}

func NewProviderAccount(user *User, provider *SocialProvider, userInfo *oidc.UserInfo) Account {
	if provider == nil {
		panic("provider can not be nil")
	}

	return Account{
		BaseModel: BaseModel{
			ID: CreateTypedId("acc", 12),
		},

		Provider:   provider,
		ProviderID: &provider.ID,

		User:   user,
		UserID: user.ID,

		Username: userInfo.Username,
		Password: nil,

		PlatformID:     userInfo.ID,
		PlatformAvatar: userInfo.ProfilePicture,
	}
}
