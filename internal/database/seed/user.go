package seed

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
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
		ID:        database.CreateID("usr"),
		Name:      "admin",
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateUser(ctx, user); err != nil {
		return err
	}

	password := cryptography.CreateSecureString(24)
	hashed, algorithm := cryptography.HashString(password)
	account := &database.Account{
		ID:          database.CreateID("acc"),
		UserID:      user.ID,
		Type:        "internal",
		Password:    &hashed,
		PasswordAlg: algorithm,
		CreatedAt:   time.Now(),
	}

	if err := star.Database.CreateAccount(ctx, account); err != nil {
		return err
	}

	role := star.Database.GetRoleByName(ctx, "superadmin")
	if role == nil {
		return errors.New("superadmin role not found")
	}

	binding := &database.RoleBinding{
		ID:        database.CreateID("rb"),
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
