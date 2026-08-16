package login

import (
	"context"
	"errors"
	"log/slog"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/state"
)

var ErrLoginsDisabled = errors.New("logins are currently disabled, you may not login at this time")

// Run any last checks on user logins before allowing them through, to be used later for bans or whatever
func Preverify(ctx context.Context, star *state.StarServer, user *database.User, account *database.Account) error {
	// If logins are disabled, prevent them from actually getting in
	if !star.Settings.AllowLogins {
		canBypass := star.Database.UserHasPermission(ctx, user.ID, permission.PermissionBypassNoLogin.Name)
		if !canBypass {
			return ErrLoginsDisabled
		}

		slog.Warn("allowing user to bypass login due to permission", "user", user.ID, "name", user.Name)
	}

	return nil
}
