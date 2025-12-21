package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

	"github.com/tubblelive/star/oidc"
)

const (
	ServiceUserID    = "service"
	ServiceUserKeyID = "key1"
)

type User struct {
	BaseModel
	Username       string    `json:"username" gorm:"uniqueIndex"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email" gorm:"uniqueIndex"`
	EmailVerified  bool      `json:"email_verified"`
	ProfilePicture string    `json:"-"`
	EmailCode      string    `json:"-"`
	Roles          []Role    `json:"-" gorm:"many2many:user_roles"`
	Accounts       []Account `json:"-"`
}

func (u *User) GenerateProfilePicture() string {
	if u.ProfilePicture != "" {
		return u.ProfilePicture
	}

	input := strings.TrimSpace(strings.ToLower(u.Email))
	hashed := sha256.Sum256([]byte(input))

	hash := hex.EncodeToString(hashed[:])

	// Use gravatar as the default pfp
	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=initials&r=pg", hash)
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

		Username: username,

		Email:         strings.ToLower(email),
		EmailVerified: false,
		EmailCode:     createEmailCode(),
	}
}

func NewUserFromInfo(userInfo *oidc.UserInfo) User {
	return User{
		BaseModel: BaseModel{
			ID: CreateTypedId("usr", 12),
		},

		Username:       userInfo.Username,
		FirstName:      userInfo.FirstName,
		LastName:       userInfo.LastName,
		ProfilePicture: userInfo.ProfilePicture,

		Email:         strings.ToLower(userInfo.Email),
		EmailVerified: false,
		EmailCode:     createEmailCode(),
	}
}
