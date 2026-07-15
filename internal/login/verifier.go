package login

import (
	"context"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

// Run any last checks on user logins before allowing them through, to be used later for bans or whatever
func Preverify(ctx context.Context, star *state.StarServer, user *database.User, account *database.Account) error {
	return nil
}
