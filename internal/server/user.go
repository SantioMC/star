package server

import (
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/audit"
	"github.com/tubblelive/star/internal/birthday"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/export"
	"github.com/tubblelive/star/internal/openid"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/permission"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/server/transformer"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/username"
)

func userRouter(router fiber.Router) {
	router.Get(
		"/",
		middleware.RequireAuth,
		middleware.RequireScope(openid.ScopeAdminUsersRead),
		middleware.RequirePermission(permission.PermissionAdminReadUsers),
		getAllUsers,
	)

	router.Get("/register", getRegisterRequirements)
	router.Post("/register", middleware.RequireCaptcha, registerUser)

	router.Post("/password/change", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserPasswordWrite), middleware.RequirePermission(permission.PermissionChangePassword), changePassword)
	router.Delete("/password", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserPasswordWrite), middleware.RequirePermission(permission.PermissionChangePassword), deletePassword)
	router.Route("/password/reset", passwordResetRouter)
	router.Post(
		"/password/require-change",
		middleware.RequireAuth,
		middleware.RequireScope(openid.ScopeAdminUsersWrite),
		middleware.RequireScope(openid.ScopeAdminTokensWrite),
		middleware.RequirePermission(permission.PermissionAdminRequirePasswordChange),
		adminRequirePasswordChangeForAll,
	)

	router.Route("/login", loginRouter)

	router.Post("/session", middleware.RateLimit(30), refreshUserSession)
	router.Delete("/session", middleware.RateLimit(30), endUserSession)

	router.Get("/me", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserRead), getUser)
	router.Patch("/me", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserWrite), middleware.RequirePermission(permission.PermissionModifySelf), patchUser)

	router.Get("/me/export", middleware.RequireAuth, middleware.RequireScope(openid.ScopeUserRead), middleware.RequirePermission(permission.PermissionExportSelf), exportSelf)
	router.Get("/:user/export", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersRead), middleware.RequirePermission(permission.PermissionAdminExportUsers), adminExportUser)

	router.Get("/:user", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersRead), middleware.RequirePermission(permission.PermissionAdminReadUsers), adminGetUser)
	router.Patch("/:user", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminUpdateUser)
	router.Delete("/:user", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersDelete), middleware.RequirePermission(permission.PermissionAdminDeleteUsers), adminDeleteUser)
	router.Get("/:user/sessions", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersRead), middleware.RequirePermission(permission.PermissionAdminReadUsers), adminGetUserSessions)
	router.Delete("/:user/sessions", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersDelete), middleware.RequirePermission(permission.PermissionAdminDeleteUsers), adminInvalidateUserSessions)
	router.Delete("/:user/sessions/:session", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersDelete), middleware.RequirePermission(permission.PermissionAdminDeleteUsers), adminInvalidateUserSession)
	router.Post("/:user/password", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminSetPassword)
	router.Post("/:user/password/require-change", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminRequirePasswordChange)
	router.Get("/:user/impersonate/applications", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersImpersonate), middleware.RequirePermission(permission.PermissionAdminImpersonateUsers), getImpersonationApplications)
	router.Post("/:user/impersonate", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersImpersonate), middleware.RequirePermission(permission.PermissionAdminImpersonateUsers), impersonateUser)
	router.Delete("/:user/lockout", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminClearLockout)
	router.Get("/:user/suspensions", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersRead), middleware.RequirePermission(permission.PermissionAdminReadUsers), adminGetUserSuspensions)
	router.Post("/:user/suspend", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminSuspendUsers), adminSuspendUser)
	router.Post("/:user/unsuspend", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminSuspendUsers), adminUnsuspendUser)
	router.Post("/:user/invite", middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminInviteUsers), regenerateInvite)
	router.Post("/:user/invite/email", middleware.RateLimit(10), middleware.RequireAuth, middleware.RequireScope(openid.ScopeAdminUsersWrite), middleware.RequirePermission(permission.PermissionAdminInviteUsers), emailInvite)
}

// misc: maybe move this to transformer?
func populateUser(ctx fiber.Ctx, user *database.User) transformer.AdminUser {
	star := common.GetState(ctx)

	email := star.Database.GetPrimaryEmail(ctx, user.ID)
	roles := star.Database.GetUserRoles(ctx, user.ID)

	reach, err := request.GetWeight(ctx)
	if err != nil {
		slog.Error("failed to get requester weight", "err", err)
		reach = database.DefaultRoleWeight
	}

	transformed := make([]transformer.AdminRole, len(roles))
	for idx, role := range roles {
		permissions := star.Database.GetRolePermissions(ctx, role.ID)
		transformed[idx] = transformer.TransformRole(role, permissions, reach)
	}

	attempts, err := star.Database.GetLoginAttempts(ctx, user.ID)
	if err != nil {
		slog.Error("failed to get login attempts", "user", user.ID, "err", err)
	}

	suspension, err := star.Database.GetActiveSuspension(ctx, user.ID)
	if err != nil {
		slog.Error("failed to get active suspension", "user", user.ID, "err", err)
	}

	invite, err := star.Database.GetUserInvite(ctx, user.ID)
	if err != nil {
		slog.Error("failed to get user invite", "user", user.ID, "err", err)
	}

	account := star.Database.GetInternalAccountByUserID(ctx, user.ID)

	return transformer.TransformAdminUser(user, email, transformed, attempts, account, suspension, invite, reach)
}

func requireManagement(ctx fiber.Ctx, user *database.User) error {
	star := common.GetState(ctx)

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	weight, err := star.Database.GetUserWeight(ctx, user.ID)
	if err != nil {
		return err
	}

	if weight > reach {
		return fiber.NewError(403, "You are not permitted to manage a user with a higher role than your own")
	}

	return nil
}

func getRegisterRequirements(ctx fiber.Ctx) error {
	policy := common.GetState(ctx).Settings.Load().GetAgePolicy()

	return ctx.JSON(types.RegisterRequirements{
		Username:         username.DefaultSettings,
		BirthdayTracked:  policy.Tracked,
		BirthdayRequired: policy.Enabled(),
	})
}

func optionalName(value *string) *string {
	if value != nil && len(*value) == 0 {
		return nil
	}

	return value
}

func registerUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	settings := star.Settings.Load()

	if !settings.AllowRegistration {
		return fiber.NewError(403, "Registration is currently disabled")
	}

	var request types.RegisterRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	request.FirstName = optionalName(request.FirstName)
	request.LastName = optionalName(request.LastName)

	if !settings.TrackBirthdays {
		request.Birthday = nil
	}

	declaredBirthday, err := birthday.Parse(request.Birthday)
	if err != nil {
		return err
	}

	if err := settings.GetAgePolicy().Check(declaredBirthday); err != nil {
		return err
	}

	username, err := username.ValidateUsername(request.Username, &username.DefaultSettings)
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	password, err := password.ValidatePassword(request.Password, settings.GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	if err := settings.GetEmailDomainPolicy().Check(request.Email); err != nil {
		return err
	}

	hashed, alg := cryptography.HashString(password)

	if star.Database.EmailClaimed(ctx, request.Email) {
		return fiber.NewError(400, "username or email already taken")
	}

	if len(star.Database.GetUsersByName(ctx, username)) > 0 {
		return fiber.NewError(400, "username or email already taken")
	}

	created := time.Now()
	user := &database.User{
		ID:        database.CreateID(database.UserKey),
		Name:      username,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Birthday:  declaredBirthday,
		CreatedAt: created,
	}

	email := newUserEmail(settings, user.ID, request.Email, true, created)

	account := &database.Account{
		ID:          database.CreateID(database.AccountKey),
		Type:        "internal",
		UserID:      user.ID,
		Email:       &request.Email,
		Name:        &username,
		Password:    &hashed,
		PasswordAlg: alg,
		CreatedAt:   created,
	}

	err = star.Database.CreateUserIdentity(ctx, user, email, account)
	if errors.Is(err, database.ErrDuplicate) {
		return fiber.NewError(400, "username or email already taken")
	} else if err != nil {
		return err
	}

	if !email.Verified {
		sendVerificationEmail(star, username, request.Email, email.Code)
	}

	audit.Log(ctx, audit.UserRegister, audit.By(user))

	return ctx.JSON(fiber.Map{
		"success":               true,
		"verification_required": !email.Verified,
	})
}

func changePassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.ChangePasswordRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	account := star.Database.GetInternalAccountByUserID(ctx, user.ID)
	if account != nil && account.Password != nil {
		if data.OldPassword == nil {
			return fiber.NewError(401, "The old password was not specified")
		}

		success := cryptography.VerifyHash(*data.OldPassword, *account.Password)
		if !success {
			return fiber.NewError(401, "The old password is incorrect")
		}
	}

	sanitized, err := password.ValidatePassword(data.NewPassword, star.Settings.Load().GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(sanitized)

	if account != nil {
		// Update if we have an internal account
		account.Password = &hashed
		account.PasswordAlg = alg
		account.PasswordChangeRequired = false

		if err := star.Database.UpdateAccount(ctx, account); err != nil {
			return err
		}
	} else {
		// Create if we have no internal account
		email := star.Database.GetPrimaryEmail(ctx, user.ID)
		if email == nil {
			return errors.New("no primary email found for user")
		}

		account = &database.Account{
			ID:          database.CreateID(database.AccountKey),
			Type:        "internal",
			UserID:      user.ID,
			Email:       &email.Email,
			Name:        &user.Name,
			Password:    &hashed,
			PasswordAlg: alg,
			CreatedAt:   time.Now(),
		}

		if err := star.Database.CreateAccount(ctx, account); err != nil {
			return err
		}
	}

	notifyPasswordChanged(ctx, star, user.ID, star.ClientIP(ctx))
	audit.Log(ctx, audit.PasswordChange)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func deletePassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	accounts := star.Database.GetAccountsByUserID(ctx, userID)
	var internalAccount *database.Account = nil
	for _, account := range accounts {
		if account.Type == "internal" {
			internalAccount = &account
			break
		}
	}

	if internalAccount == nil {
		return fiber.NewError(400, "You do not have a password set")
	}

	removable, err := star.Database.CanRemoveAccount(ctx, internalAccount)
	if err != nil {
		return err
	} else if !removable {
		return fiber.NewError(400, "You must add another form of authentication before deleting your password")
	}

	if err := star.Database.DeleteAccountByID(ctx, internalAccount.ID); err != nil {
		return err
	}

	audit.Log(ctx, audit.PasswordRemove)

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func getUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	emails := star.Database.GetEmailsByUserID(ctx, user.ID)
	roles := []database.Role{}
	permissions := []database.Permission{}
	sessions := []database.Session{}
	accounts := []database.Account{}

	if request.HasScope(ctx, openid.ScopeUserRolesRead) {
		roles = star.Database.GetUserRoles(ctx, user.ID)
		permissions = star.Database.GetUserPermissions(ctx, user.ID)
	}

	if request.HasScope(ctx, openid.ScopeUserSessionsRead) {
		sessions = star.Database.GetUserSessions(ctx, user.ID)
	}

	if request.HasScope(ctx, openid.ScopeUserAccountRead) {
		accounts = star.Database.GetAccountsByUserID(ctx, user.ID)
	}

	response := transformer.TransformUser(
		user,
		emails,
		roles,
		permissions,
		sessions,
		accounts,
		request.HasScope(ctx, openid.ScopeUserEmailRead),
		star.Settings.Load().TrackBirthdays && request.HasScope(ctx, openid.ScopeUserBirthdayRead),
	)

	return ctx.JSON(response)
}

func patchUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.UpdateUserRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	before := *user

	if data.Username != nil && *data.Username != user.Name {
		sanitized, err := username.ValidateUsername(*data.Username, &username.DefaultSettings)
		if err != nil {
			return fiber.NewError(400, err.Error())
		}

		for _, existing := range star.Database.GetUsersByName(ctx, sanitized) {
			if existing.ID != user.ID {
				return fiber.NewError(400, "This username is already taken")
			}
		}

		user.Name = sanitized
	}

	if data.FirstName != nil {
		user.FirstName = data.FirstName
	}

	if data.LastName != nil {
		user.LastName = data.LastName
	}

	if data.Birthday != nil && star.Settings.Load().TrackBirthdays {
		user.Birthday, err = birthday.Parse(data.Birthday)
		if err != nil {
			return err
		}

		if err := star.Settings.Load().GetAgePolicy().Check(user.Birthday); err != nil {
			return err
		}
	}

	if err := star.Database.UpdateUser(ctx, user); err != nil {
		return err
	}

	audit.Log(ctx, audit.UserUpdate, audit.Changes(before, *user))

	return ctx.JSON(populateUser(ctx, user))
}

func getAllUsers(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	page, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	query := ctx.Query("query")

	var users []database.User
	if query != "" {
		users = star.Database.SearchUsers(ctx, query, page, limit)
	} else {
		users = star.Database.GetUsers(ctx, page, limit)
	}

	transformed := make([]transformer.AdminUser, len(users))
	for idx := range users {
		transformed[idx] = populateUser(ctx, &users[idx])
	}

	return ctx.JSON(transformed)
}

func adminGetUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	return ctx.JSON(populateUser(ctx, user))
}

func exportSelf(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	userID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, userID)
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	return sendExport(ctx, user)
}

func adminExportUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	return sendExport(ctx, user)
}

func sendExport(ctx fiber.Ctx, user *database.User) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	slog.Info("user data exported", "user", user.ID, "requester", requesterID)
	audit.Log(ctx, audit.UserExport, audit.OnUser(user))

	ctx.Set(fiber.HeaderCacheControl, "no-store")
	ctx.Attachment("star-export-" + user.Name + "-" + time.Now().UTC().Format(time.DateOnly) + ".json")

	exported, err := export.User(ctx, star.Database, user)
	if err != nil {
		return err
	}

	return ctx.JSON(exported)
}

func adminUpdateUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.AdminUpdateUserRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	before := *user

	if data.Username != nil && *data.Username != user.Name {
		sanitized, err := username.ValidateUsername(*data.Username, &username.DefaultSettings)
		if err != nil {
			return fiber.NewError(400, err.Error())
		}

		for _, existing := range star.Database.GetUsersByName(ctx, sanitized) {
			if existing.ID != user.ID {
				return fiber.NewError(400, "This username is already taken")
			}
		}

		user.Name = sanitized
	}

	if data.FirstName != nil {
		user.FirstName = data.FirstName
	}

	if data.LastName != nil {
		user.LastName = data.LastName
	}

	if data.Birthday != nil && star.Settings.Load().TrackBirthdays {
		parsed, err := birthday.Parse(data.Birthday)
		if err != nil {
			return err
		}

		user.Birthday = parsed
	}

	if err := star.Database.UpdateUser(ctx, user); err != nil {
		return err
	}

	if data.Username != nil || data.FirstName != nil || data.LastName != nil || data.Birthday != nil {
		audit.Log(ctx, audit.UserUpdate, audit.OnUser(user), audit.Changes(before, *user))
	}

	if data.Roles != nil {
		if err := setUserRoles(ctx, user, *data.Roles); err != nil {
			return err
		}
	}

	return ctx.JSON(populateUser(ctx, user))
}

func setUserRoles(ctx fiber.Ctx, user *database.User, roleIDs []string) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	if requesterID == user.ID {
		return fiber.NewError(403, "You are not permitted to modify your own roles")
	}

	reach, err := request.GetWeight(ctx)
	if err != nil {
		return err
	}

	requested := slices.Compact(slices.Sorted(slices.Values(roleIDs)))
	for _, roleID := range requested {
		role := star.Database.GetRoleByID(ctx, roleID)
		if role == nil {
			return fiber.NewError(400, "Unknown role: "+roleID)
		} else if role.IsDefault() {
			return fiber.NewError(400, "The default role is already applied to every user")
		} else if role.Weight > reach {
			return fiber.NewError(403, "You are not permitted to assign a role higher than your own")
		}
	}

	changed, err := star.Database.SetUserRoles(ctx, user.ID, requested)
	if err != nil || !changed {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.UserRolesUpdate, audit.OnUser(user), audit.With("roles", requested))

	return nil
}

func adminSetPassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if requesterID == user.ID {
		return fiber.NewError(403, "Use the password change endpoint to update your own password")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	invite, err := star.Database.GetUserInvite(ctx, user.ID)
	if err != nil {
		return err
	} else if invite != nil {
		return fiber.NewError(400, "This user has a pending invite and sets their own password when accepting it")
	}

	var data types.AdminSetPasswordRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	sanitized, err := password.ValidatePassword(data.Password, star.Settings.Load().GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(sanitized)

	account := star.Database.GetInternalAccountByUserID(ctx, user.ID)
	if account == nil {
		email := star.Database.GetPrimaryEmail(ctx, user.ID)
		if email == nil {
			return fiber.NewError(400, "This user has no primary email to create an internal account with")
		}

		account = &database.Account{
			ID:          database.CreateID(database.AccountKey),
			Type:        "internal",
			UserID:      user.ID,
			Email:       &email.Email,
			Name:        &user.Name,
			Password:    &hashed,
			PasswordAlg: alg,
			CreatedAt:   time.Now(),
		}

		if err := star.Database.CreateAccount(ctx, account); err != nil {
			return err
		}
	} else {
		account.Password = &hashed
		account.PasswordAlg = alg

		if err := star.Database.UpdateAccount(ctx, account); err != nil {
			return err
		}
	}

	if err := star.Database.RevokeUserSessions(ctx, user.ID, database.RevocationReasonPasswordReset); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	notifyPasswordChanged(ctx, star, user.ID, "")
	audit.Log(ctx, audit.PasswordSet, audit.OnUser(user))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func adminRequirePasswordChange(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if requesterID == user.ID {
		return fiber.NewError(403, "Use the password change endpoint to update your own password")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	required, err := star.Database.RequirePasswordChange(ctx, user.ID)
	if err != nil {
		return err
	} else if !required {
		return fiber.NewError(400, "This user does not have a password to change")
	}

	cache.UpdateRevocations(ctx, star)
	slog.Info("password change required", "user", user.ID, "by", requesterID)
	audit.Log(ctx, audit.PasswordRequireChange, audit.OnUser(user))

	return ctx.JSON(populateUser(ctx, user))
}

func adminRequirePasswordChangeForAll(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	affected, err := star.Database.RequirePasswordChangeForAll(ctx)
	if err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	slog.Warn("password change required for all users", "affected", affected, "by", requesterID)
	audit.Log(ctx, audit.PasswordRequireAll, audit.With("affected", affected))

	return ctx.JSON(fiber.Map{
		"affected": affected,
	})
}

func adminDeleteUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if user.ID == requesterID {
		return fiber.NewError(403, "You are not permitted to delete your own user")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	if err := star.Database.DeleteUser(ctx, user.ID); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.UserDelete, audit.OnUser(user))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func adminClearLockout(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserID(ctx)
	if err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	if err := star.Database.ClearLoginAttempts(ctx, user.ID); err != nil {
		return err
	}

	slog.Info("login lockout cleared", "user", user.ID, "by", requesterID)
	audit.Log(ctx, audit.UserLockoutClear, audit.OnUser(user))

	return ctx.JSON(populateUser(ctx, user))
}

func adminGetUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	_, limit, err := request.GetPagination(ctx)
	if err != nil {
		return err
	}

	before, err := request.GetCursor(ctx)
	if err != nil {
		return err
	}

	sessions, err := star.Database.GetActiveUserSessions(ctx, user.ID, before, limit+1)
	if err != nil {
		return err
	}

	page := types.NewPage(sessions, limit, func(last database.Session) database.Cursor {
		return database.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	})

	return ctx.JSON(types.Page[transformer.UserSession]{
		Items: transformer.TransformSessions(page.Items),
		Next:  page.Next,
	})
}

func adminInvalidateUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	if err := star.Database.RevokeUserSessions(ctx, user.ID, database.RevocationReasonSessionsCleared); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.SessionRevokeAll, audit.OnUser(user))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func adminInvalidateUserSession(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := requireManagement(ctx, user); err != nil {
		return err
	}

	session := star.Database.GetSessionByFamily(ctx, ctx.Params("session"))
	if session == nil || session.UserID != user.ID {
		return fiber.NewError(404, "Session not found")
	}

	if err := star.Database.RevokeSessionFamily(ctx, user.ID, session.FamilyID, database.RevocationReasonSignedOut); err != nil {
		return err
	}

	cache.UpdateRevocations(ctx, star)
	audit.Log(ctx, audit.SessionRevoke, audit.OnUser(user), audit.With("session", session.FamilyID))

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
