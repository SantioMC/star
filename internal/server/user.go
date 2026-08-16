package server

import (
	"errors"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/emails"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/email"
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
		middleware.RequirePermission(permission.PermissionAdminReadUsers),
		getAllUsers,
	)

	router.Post("/register", middleware.RequireCaptcha, registerUser)

	router.Post("/password/change", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionChangePassword), changePassword)
	router.Delete("/password", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionChangePassword), deletePassword)

	router.Route("/login", loginRouter)

	router.Get("/:user", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAdminReadUsers), adminGetUser)
	router.Patch("/:user", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminUpdateUser)
	router.Delete("/:user", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAdminDeleteUsers), adminDeleteUser)
	router.Delete("/:user/sessions", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAdminDeleteUsers), adminInvalidateUserSessions)
	router.Post("/:user/password", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAdminWriteUsers), adminSetPassword)
}

// misc: maybe move this to transformer?
func populateUser(ctx fiber.Ctx, user *database.User) transformer.AdminUser {
	star := common.GetState(ctx)

	email := star.Database.GetPrimaryEmail(ctx, user.ID)
	roles := star.Database.GetUserRoles(ctx, user.ID)

	transformed := make([]transformer.AdminRole, len(roles))
	for idx, role := range roles {
		permissions := star.Database.GetRolePermissions(ctx, role.ID)
		transformed[idx] = transformer.TransformRole(role, permissions)
	}

	return transformer.TransformAdminUser(user, email, transformed)
}

func registerUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	if !star.Settings.AllowRegistration {
		return fiber.NewError(403, "Registration is currently disabled")
	}

	var request types.RegisterRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if request.FirstName != nil && len(*request.FirstName) == 0 {
		request.FirstName = nil
	}

	if request.LastName != nil && len(*request.LastName) == 0 {
		request.LastName = nil
	}

	username, err := username.ValidateUsername(request.Username, &username.DefaultSettings)
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	password, err := password.ValidatePassword(request.Password, star.Settings.GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(password)

	if star.Database.EmailExists(ctx, request.Email) {
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
		CreatedAt: created,
	}

	if err := star.Database.CreateUser(ctx, user); err != nil {
		return err
	}

	email := &database.UserEmail{
		ID:        database.CreateID(database.EmailKey),
		UserID:    user.ID,
		Email:     request.Email,
		Verified:  false,
		Primary:   true,
		Code:      email.CreateCode(star.Settings.EmailCodeLength),
		CreatedAt: created,
	}

	if err := star.Database.CreateEmail(ctx, email); err != nil {
		return err
	}

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

	if err := star.Database.CreateAccount(ctx, account); err != nil {
		return err
	}

	// Send verification email
	values := url.Values{}
	values.Set("email", request.Email)
	values.Set("code", email.Code)

	url := star.Config.ServerURL + "/verify?" + values.Encode()
	if err := star.Email.Send(
		request.Email,
		"Verify your email address",
		emails.VerificationEmail,
		map[string]string{
			"Name":            username,
			"VerificationURL": url,
			"Code":            email.Code,
		},
	); err != nil {
		slog.Warn("failed to send verification email", "err", err, "email", request.Email, "code", email.Code)
	}

	return ctx.JSON(fiber.Map{
		"success": true,
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

	sanitized, err := password.ValidatePassword(data.NewPassword, star.Settings.GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	hashed, alg := cryptography.HashString(sanitized)

	if account != nil {
		// Update if we have an internal account
		account.Password = &hashed
		account.PasswordAlg = alg

		if err := star.Database.UpdateAccount(ctx, account); err != nil {
			return err
		}

		return ctx.JSON(fiber.Map{
			"success": true,
		})
	}

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

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func deletePassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	userId, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	accounts := star.Database.GetAccountsByUserID(ctx, userId)
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

	if len(accounts) <= 1 {
		return fiber.NewError(400, "You must add another form of authentication before deleting your password")
	}

	if err := star.Database.DeleteAccountByID(ctx, internalAccount.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func getAllUsers(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "25"))
	if err != nil {
		return err
	}

	if page < 1 {
		return fiber.NewError(400, "page must be at least 1")
	}

	if limit < 1 || limit > 100 {
		return fiber.NewError(400, "limit must be between 1 and 100")
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

	if err := star.Database.UpdateUser(ctx, user); err != nil {
		return err
	}

	if data.Roles != nil {
		if err := setUserRoles(ctx, user.ID, *data.Roles); err != nil {
			return err
		}
	}

	return ctx.JSON(populateUser(ctx, user))
}

func setUserRoles(ctx fiber.Ctx, userID string, roleIDs []string) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserId(ctx)
	if err != nil {
		return err
	}

	if requesterID == userID {
		return fiber.NewError(403, "You are not permitted to modify your own roles")
	}

	requested := slices.Compact(slices.Sorted(slices.Values(roleIDs)))
	for _, roleID := range requested {
		if star.Database.GetRoleByID(ctx, roleID) == nil {
			return fiber.NewError(400, "Unknown role: "+roleID)
		}
	}

	current := star.Database.GetUserRoleBindings(ctx, userID)
	for _, binding := range current {
		if !slices.Contains(requested, binding.RoleID) {
			if err := star.Database.DeleteRoleBindingByID(ctx, binding.ID); err != nil {
				return err
			}
		}
	}

	for _, roleID := range requested {
		exists := slices.ContainsFunc(current, func(binding database.RoleBinding) bool {
			return binding.RoleID == roleID
		})

		if exists {
			continue
		}

		binding := &database.RoleBinding{
			ID:        database.CreateID(database.RoleBindingKey),
			RoleID:    roleID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}

		if err := star.Database.CreateRoleBinding(ctx, binding); err != nil {
			return err
		}
	}

	return nil
}

func adminSetPassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserId(ctx)
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

	var data types.AdminSetPasswordRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	sanitized, err := password.ValidatePassword(data.Password, star.Settings.GetPasswordSettings())
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

	if err := star.Database.DeleteUserSessions(ctx, user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func adminDeleteUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	requesterID, err := request.GetUserId(ctx)
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

	if err := star.Database.DeleteUser(ctx, user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func adminInvalidateUserSessions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user := star.Database.GetUserByID(ctx, ctx.Params("user"))
	if user == nil {
		return fiber.NewError(404, "User not found")
	}

	if err := star.Database.DeleteUserSessions(ctx, user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
