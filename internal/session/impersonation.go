package session

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/state"
)

var ErrImpersonationDenied = fiber.NewError(403, "You are not permitted to impersonate this user")

func CheckImpersonation(ctx fiber.Ctx, server *state.StarServer, impersonatorID string, userID string) error {
	if impersonatorID == userID {
		return ErrImpersonationDenied
	}

	if err := rejectSuspended(ctx, server, impersonatorID); err != nil {
		return err
	}

	if err := rejectSuspended(ctx, server, userID); err != nil {
		return err
	}

	if !server.Database.UserHasPermission(ctx, impersonatorID, permission.PermissionAdminImpersonateUsers.Name) {
		return ErrImpersonationDenied
	}

	impersonatorWeight, err := server.Database.GetUserWeight(ctx, impersonatorID)
	if err != nil {
		return err
	}

	userWeight, err := server.Database.GetUserWeight(ctx, userID)
	if err != nil {
		return err
	}

	if impersonatorWeight < userWeight {
		return ErrImpersonationDenied
	}

	return nil
}
