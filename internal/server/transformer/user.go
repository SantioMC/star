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

type PublicUser struct {
	ID             string        `json:"id"`
	Username       string        `json:"username"`
	FirstName      *string       `json:"first_name"`
	LastName       *string       `json:"last_name"`
	ProfilePicture string        `json:"picture"`
	Email          *string       `json:"email,omitempty"`
	EmailVerified  bool          `json:"email_verified,omitempty"`
	Roles          []string      `json:"roles,omitempty"`
	Permissions    []string      `json:"permissions,omitempty"`
	Sessions       []UserSession `json:"sessions,omitempty"`
	Accounts       []UserAccount `json:"accounts,omitempty"`
	Emails         []Email       `json:"emails,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
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
	Lockout        *Lockout    `json:"lockout"`
	CreatedAt      time.Time   `json:"created_at"`
}

type Lockout struct {
	Locked        bool       `json:"locked"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt time.Time  `json:"last_attempt_at"`
	LockedUntil   *time.Time `json:"locked_until"`
}

func TransformLockout(attempts *database.LoginAttempts) *Lockout {
	if attempts == nil {
		return nil
	}

	return &Lockout{
		Locked:        attempts.IsLocked(),
		Attempts:      attempts.Attempts,
		LastAttemptAt: attempts.LastAttemptAt,
		LockedUntil:   attempts.LockedUntil,
	}
}

func TransformAdminUser(
	user *database.User,
	email *database.UserEmail,
	roles []AdminRole,
	attempts *database.LoginAttempts,
) AdminUser {
	var primaryEmail *string

	if email != nil {
		primaryEmail = &email.Email
	}

	return AdminUser{
		ID:             user.ID,
		Username:       user.Name,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		ProfilePicture: picture.GetProfilePicture(user, primaryEmail),
		Email:          primaryEmail,
		EmailVerified:  email != nil && email.Verified,
		Roles:          roles,
		Lockout:        TransformLockout(attempts),
		CreatedAt:      user.CreatedAt,
	}
}

func TransformUser(
	user *database.User,
	emails []database.UserEmail,
	roles []database.Role,
	permissions []database.Permission,
	sessions []database.Session,
	accounts []database.Account,
	includeEmail bool,
) PublicUser {
	primaryEmail := ""
	emailVerified := false

	for _, email := range emails {
		if email.Primary {
			primaryEmail = email.Email
			emailVerified = email.Verified
		}
	}

	roleNames := []string{}
	permissionNames := []string{}
	userSessions := TransformSessions(sessions)

	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
	}

	transformed := PublicUser{
		ID:             user.ID,
		Username:       user.Name,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		ProfilePicture: picture.GetProfilePicture(user, &primaryEmail),
		Accounts:       TransformAccounts(accounts, includeEmail),
		Roles:          roleNames,
		Permissions:    permissionNames,
		Sessions:       userSessions,
		CreatedAt:      user.CreatedAt,
	}

	if includeEmail {
		transformed.Email = &primaryEmail
		transformed.EmailVerified = emailVerified
		transformed.Emails = TransformEmails(emails)
	}

	return transformed
}
