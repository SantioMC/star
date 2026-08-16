package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/picture"
)

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

type AdminUser struct {
	ID             string      `json:"id"`
	Username       string      `json:"username"`
	FirstName      *string     `json:"first_name"`
	LastName       *string     `json:"last_name"`
	ProfilePicture string      `json:"profile_picture"`
	Email          *string     `json:"email"`
	EmailVerified  bool        `json:"email_verified"`
	Roles          []AdminRole `json:"roles"`
	CreatedAt      time.Time   `json:"created_at"`
}

func TransformAdminUser(
	user *database.User,
	email *database.UserEmail,
	roles []AdminRole,
) AdminUser {
	effectiveEmail := ""
	var primaryEmail *string

	if email != nil {
		effectiveEmail = email.Email
		primaryEmail = &email.Email
	}

	return AdminUser{
		ID:             user.ID,
		Username:       user.Name,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		ProfilePicture: picture.GetProfilePicture(user, effectiveEmail),
		Email:          primaryEmail,
		EmailVerified:  email != nil && email.Verified,
		Roles:          roles,
		CreatedAt:      user.CreatedAt,
	}
}
