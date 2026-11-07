package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/models"
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
	Birthday       *models.Date  `json:"birthday,omitempty"`
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
	ID                     string                   `json:"id"`
	Username               string                   `json:"username"`
	Type                   database.UserType        `json:"type"`
	FirstName              *string                  `json:"first_name"`
	LastName               *string                  `json:"last_name"`
	ProfilePicture         string                   `json:"profile_picture"`
	Birthday               *models.Date             `json:"birthday"`
	Email                  *string                  `json:"email"`
	EmailVerified          bool                     `json:"email_verified"`
	Roles                  []AdminRole              `json:"roles"`
	Lockout                *Lockout                 `json:"lockout"`
	Suspension             *database.UserSuspension `json:"suspension"`
	Invite                 *database.UserInvite     `json:"invite"`
	HasPassword            bool                     `json:"has_password"`
	PasswordChangeRequired bool                     `json:"password_change_required"`
	Manageable             bool                     `json:"manageable"`
	CreatedAt              time.Time                `json:"created_at"`
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
	account *database.Account,
	suspension *database.UserSuspension,
	invite *database.UserInvite,
	reach int,
) AdminUser {
	var primaryEmail *string

	if email != nil {
		primaryEmail = &email.Email
	}

	weight := database.DefaultRoleWeight
	for _, role := range roles {
		weight = max(weight, role.Weight)
	}

	return AdminUser{
		ID:                     user.ID,
		Username:               user.Name,
		Type:                   user.Type,
		FirstName:              user.FirstName,
		LastName:               user.LastName,
		ProfilePicture:         picture.GetProfilePicture(user, primaryEmail),
		Birthday:               user.Birthday,
		Email:                  primaryEmail,
		EmailVerified:          email != nil && email.Verified,
		Roles:                  roles,
		Lockout:                TransformLockout(attempts),
		Suspension:             suspension,
		Invite:                 invite,
		HasPassword:            account != nil && account.Password != nil,
		PasswordChangeRequired: account != nil && account.PasswordChangeRequired,
		Manageable:             weight <= reach,
		CreatedAt:              user.CreatedAt,
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
	includeBirthday bool,
) PublicUser {
	primaryEmail := ""
	emailVerified := false

	for _, email := range emails {
		if email.Primary {
			primaryEmail = email.Email
			emailVerified = email.Trusted()
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

	if includeBirthday {
		transformed.Birthday = user.Birthday
	}

	return transformed
}
