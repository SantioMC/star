package login

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/state"
)

var ErrLoginsDisabled = fiber.NewError(403, "Logins are currently disabled, you may not login at this time")

// Run any last checks on user logins before allowing them through
func Preverify(ctx context.Context, star *state.StarServer, user *database.User, account *database.Account) error {
	// If logins are disabled, prevent them from actually getting in
	if !star.Settings.Load().AllowLogins {
		canBypass := star.Database.UserHasPermission(ctx, user.ID, permission.PermissionBypassNoLogin.Name)
		if !canBypass {
			return ErrLoginsDisabled
		}

		slog.Warn("allowing user to bypass login due to permission", "user", user.ID, "name", user.Name)
	}

	return nil
}
