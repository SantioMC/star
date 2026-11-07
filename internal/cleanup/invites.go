package cleanup

import (
	"context"
	"log/slog"

	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/state"
)

func CleanupExpiredInvites(ctx context.Context, star *state.StarServer) {
	users, err := star.Database.DeleteExpiredInvitedUsers(ctx)
	if err != nil {
		slog.Error("failed to delete expired invited users", "err", err)
		return
	}

	for idx := range users {
		slog.Info("invite expired", "user", users[idx].ID)
		audit.LogSystem(ctx, star, audit.UserInviteExpire, audit.OnUser(&users[idx]))
	}
}
