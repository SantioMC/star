package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/picture"
)

func TransformOAuthUser(
	user *database.User,
	email *database.UserEmail,
) OAuthUser {
	effectiveEmail := ""
	if email != nil {
		effectiveEmail = email.Email
	}

	firstName := ""
	lastName := ""

	if user.FirstName != nil {
		firstName = *user.FirstName
	}

	if user.LastName != nil {
		lastName = *user.LastName
	}

	profilePicture := picture.GetProfilePicture(user, effectiveEmail)

	return OAuthUser{
		ID:                user.ID,
		Name:              user.Name,
		Email:             effectiveEmail,
		EmailVerified:     email != nil && email.Verified,
		FirstName:         firstName,
		LastName:          lastName,
		Nickname:          user.Name,
		PreferredUsername: user.Name,
		ProfilePicture:    profilePicture,
	}
}

func TransformEmail(email *database.UserEmail) Email {
	return Email{
		ID:        email.ID,
		Email:     email.Email,
		Verified:  email.Verified,
		Primary:   email.Primary,
		CreatedAt: email.CreatedAt,
	}
}

func TransformEmails(emails []database.UserEmail) []Email {
	transformed := make([]Email, len(emails))
	for i, email := range emails {
		transformed[i] = TransformEmail(&email)
	}
	return transformed
}

type Email struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	Primary   bool      `json:"primary"`
	CreatedAt time.Time `json:"created_at"`
}

type OAuthUser struct {
	ID                string `json:"sub"`
	Name              string `json:"name,omitempty"`
	FirstName         string `json:"given_name,omitempty"`
	LastName          string `json:"family_name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	ProfilePicture    string `json:"picture,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Gender            string `json:"gender,omitempty"`
}
