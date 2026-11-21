package server

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/birthday"
	"github.com/tubblelive/star/internal/branding"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

var ErrInviteInvalid = fiber.NewError(410, "This invite link is invalid or has expired, please ask for a new one")

func inviteRouter(router fiber.Router) {
	limiter := middleware.RateLimit(10)

	router.Post("/", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminInviteUsers), createInvite)
	router.Post("/lookup", limiter, lookupInvite)
	router.Post("/accept", limiter, acceptInvite)
}

func inviteExpiry(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}

func newInviteLink(star *state.StarServer, invite *database.UserInvite) types.InviteLink {
	return types.InviteLink{
		UserID:       invite.UserID,
		Code:         invite.Token,
		URL:          emails.InviteURL(star.Config.ServerURL, invite.Token),
		ExpiresAt:    invite.ExpiresAt,
		EmailEnabled: star.EmailConfig().IsEnabled(),
	}
}

func createInvite(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	settings := star.Settings.Load()

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	var data types.CreateInviteRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	created := time.Now()
	user, email, err := newManagedUser(settings, &data.UserProfileRequest, created)
	if err != nil {
		return err
	}

	invite := &database.UserInvite{
		ID:        database.CreateID(database.UserInviteKey),
		UserID:    user.ID,
		InvitedBy: &requesterID,
		Token:     database.CreateUserInviteToken(),
		ExpiresAt: inviteExpiry(data.ExpiresIn),
		CreatedAt: created,
	}

	err = star.Database.CreateInvitedUser(ctx, user, email, invite)
	if errors.Is(err, database.ErrDuplicate) {
		return fiber.NewError(400, "username or email already taken")
	} else if err != nil {
		return err
	}

	slog.Info("user invited", "user", user.ID, "by", requesterID, "expires_at", invite.ExpiresAt)
	audit.Log(ctx, audit.UserInvite, audit.OnUser(user), audit.With("expires_at", invite.ExpiresAt), audit.Private("email", email.Email))

	return ctx.JSON(newInviteLink(star, invite))
}

func pendingInviteUser(ctx fiber.Ctx) (*database.User, *database.UserInvite, error) {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return nil, nil, fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return nil, nil, err
	}

	invite, err := star.Database.GetUserInvite(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	} else if invite == nil {
		return nil, nil, fiber.NewError(400, "This user does not have a pending invite")
	}

	return user, invite, nil
}

func regenerateInvite(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user, invite, err := pendingInviteUser(ctx)
	if err != nil {
		return err
	}

	var data types.RegenerateInviteRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	invite.Token = database.CreateUserInviteToken()
	invite.ExpiresAt = inviteExpiry(data.ExpiresIn)
	invite.InvitedBy = &requesterID

	regenerated, err := star.Database.RegenerateUserInvite(ctx, invite)
	if err != nil {
		return err
	} else if !regenerated {
		return fiber.NewError(400, "This user does not have a pending invite")
	}

	slog.Info("invite regenerated", "user", user.ID, "by", requesterID, "expires_at", invite.ExpiresAt)
	audit.Log(ctx, audit.UserInviteRegenerate, audit.OnUser(user), audit.With("expires_at", invite.ExpiresAt))

	return ctx.JSON(newInviteLink(star, invite))
}

func emailInvite(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	if !star.EmailConfig().IsEnabled() {
		return fiber.NewError(400, "Email is not configured on this server")
	}

	inviter, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	user, invite, err := pendingInviteUser(ctx)
	if err != nil {
		return err
	}

	var data types.EmailInviteRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	matched := star.Database.GetUserInviteByToken(ctx, data.Code)
	if matched == nil || matched.ID != invite.ID {
		return ErrInviteInvalid
	}

	primary := star.Database.GetPrimaryEmail(ctx, user.ID)
	if primary == nil {
		return fiber.NewError(400, "This user has no primary email to send the invite to")
	}

	brand := star.Branding.Load()

	err = star.Email().Send(
		primary.Email,
		"You have been invited to "+brand.Name,
		emails.Invite,
		emails.InviteData{
			Brand:     branding.ForEmail(brand, star.Config.ServerURL),
			Name:      user.Name,
			Inviter:   inviter.Name,
			InviteURL: emails.InviteURL(star.Config.ServerURL, data.Code),
			ExpiresAt: emails.FormatTime(matched.ExpiresAt),
		},
	)

	if err != nil {
		slog.Warn("failed to send invite email", "err", err, "user", user.ID)
		return fiber.NewError(502, "Failed to send the invite email, share the link instead")
	}

	audit.Log(ctx, audit.UserInviteEmail, audit.OnUser(user), audit.Private("email", primary.Email))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func lookupInvite(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.LookupInviteRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	invite := star.Database.GetUserInviteByToken(ctx, data.Code)
	if invite == nil {
		return ErrInviteInvalid
	}

	user := star.Database.GetUserByID(ctx, invite.UserID)
	primary := star.Database.GetPrimaryEmail(ctx, invite.UserID)

	if user == nil || primary == nil {
		return ErrInviteInvalid
	}

	policy := star.Settings.Load().GetAgePolicy()

	return ctx.JSON(types.InviteDetails{
		Username:         user.Name,
		Email:            primary.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Birthday:         user.Birthday,
		BirthdayTracked:  policy.Tracked,
		BirthdayRequired: policy.Enabled(),
		ExpiresAt:        invite.ExpiresAt,
	})
}

func acceptInvite(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	settings := star.Settings.Load()

	var data types.AcceptInviteRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	invite := star.Database.GetUserInviteByToken(ctx, data.Code)
	if invite == nil {
		return ErrInviteInvalid
	}

	user := star.Database.GetUserByID(ctx, invite.UserID)
	primary := star.Database.GetPrimaryEmail(ctx, invite.UserID)

	if user == nil || primary == nil {
		return ErrInviteInvalid
	}

	if user.FirstName == nil {
		user.FirstName = optionalName(data.FirstName)
	}

	if user.LastName == nil {
		user.LastName = optionalName(data.LastName)
	}

	if user.Birthday == nil && settings.TrackBirthdays {
		declaredBirthday, err := birthday.Parse(data.Birthday)
		if err != nil {
			return err
		}

		user.Birthday = declaredBirthday
	}

	if err := settings.GetAgePolicy().Check(user.Birthday); err != nil {
		return err
	}

	if err := settings.GetEmailDomainPolicy().Check(primary.Email); err != nil {
		return err
	}

	sanitized, err := password.ValidatePassword(data.Password, settings.GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(sanitized)

	account := &database.Account{
		ID:          database.CreateID(database.AccountKey),
		Type:        "internal",
		UserID:      user.ID,
		Email:       &primary.Email,
		Name:        &user.Name,
		Password:    &hashed,
		PasswordAlg: alg,
		CreatedAt:   time.Now(),
	}

	accepted, err := star.Database.AcceptUserInvite(ctx, invite, user, account)
	if err != nil {
		return err
	} else if !accepted {
		return ErrInviteInvalid
	}

	slog.Info("invite accepted", "user", user.ID)
	audit.Log(ctx, audit.UserInviteAccept, audit.By(user))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
