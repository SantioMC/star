package seed

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
	"github.com/tubblelive/star/internal/state"
)

// Create a default admin user with full privileges
func seedUser(ctx context.Context, star *state.StarServer) error {
	users, err := star.Database.GetUserCount(ctx)
	if err != nil {
		return err
	}

	if users > 0 {
		return nil
	}

	user := &database.User{
		ID:        database.CreateID(database.UserKey),
		Name:      "admin",
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateUser(ctx, user); err != nil {
		return err
	}

	password := cryptography.CreateSecureString(24)
	hashed, algorithm := cryptography.HashString(password)
	adminEmail := "admin@star.internal"
	account := &database.Account{
		ID:          database.CreateID(database.AccountKey),
		UserID:      user.ID,
		Type:        "internal",
		Email:       &adminEmail,
		Password:    &hashed,
		PasswordAlg: algorithm,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreateAccount(ctx, account); err != nil {
		return err
	}

	userEmail := &database.UserEmail{
		ID:        database.CreateID(database.EmailKey),
		UserID:    user.ID,
		Email:     "admin@star.internal",
		Verified:  true,
		Primary:   true,
		Code:      email.CreateCode(6),
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateEmail(ctx, userEmail); err != nil {
		return err
	}

	role := star.Database.GetRoleByName(ctx, "superadmin")
	if role == nil {
		return errors.New("superadmin role not found")
	}

	binding := &database.RoleBinding{
		ID:        database.CreateID(database.RoleBindingKey),
		RoleID:    role.ID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateRoleBinding(ctx, binding); err != nil {
		return err
	}

	slog.Info("default admin account created", "username", user.Name, "password", password)
	return nil
}
