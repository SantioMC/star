package server

import (
	"errors"
	"log/slog"
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
	"github.com/tubblelive/star/internal/session"
	"github.com/tubblelive/star/internal/token"
	"github.com/tubblelive/star/internal/username"
)

func userRouter(router fiber.Router) {
	router.Post("/login", middleware.RequireCaptcha, loginUser)
	router.Post("/register", middleware.RequireCaptcha, registerUser)

	router.Post("/password/change", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionChangePassword), changePassword)
	router.Delete("/password", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionChangePassword), deletePassword)

	router.Get("/email", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionReadEmails), getUserEmails)
	router.Post("/email", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionAddEmails), addUserEmail)
	router.Delete("/email/:email", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionDeleteEmails), deleteUserEmail)
	router.Post("/email/:email/primary", middleware.RequireAuth, middleware.RequirePermission(permission.PermissionUpdatePrimaryEmail), updatePrimaryEmail)
}

func loginUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	user := star.Database.GetUserByName(ctx, request.Username)
	if user == nil {
		slog.Warn("login attempt failed", "reason", "invalid username")
		return fiber.NewError(401, "Invalid username or password")
	}

	account := star.Database.GetInternalAccountByUserID(ctx, user.ID)
	if account == nil || account.Password == nil {
		slog.Warn("login attempt failed", "reason", "no internal account")
		return fiber.NewError(401, "Invalid username or password")
	}

	success := cryptography.VerifyHash(request.Password, *account.Password)
	if !success {
		slog.Warn("login attempt failed", "reason", "invalid password")
		return fiber.NewError(401, "Invalid username or password")
	}

	permissions := []string{}
	for _, permission := range star.Database.GetUserPermissions(ctx, user.ID) {
		permissions = append(permissions, permission.Name)
	}

	session, err := session.CreateSession(
		ctx,
		star,
		&session.Options{
			User:     user,
			Account:  account,
			Duration: time.Hour * 24 * 7,
		},
	)

	if err != nil {
		return err
	}

	accessToken, err := token.Create(ctx, star, session, &token.Options{
		Duration:    time.Hour,
		Permissions: permissions,
		SuperUser:   true,
	})

	return ctx.JSON(fiber.Map{
		"next":          "complete",
		"access_token":  accessToken,
		"refresh_token": session.Token,
	})
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
		return fiber.NewError(400, "email already taken")
	}

	existing := star.Database.GetUserByName(ctx, username)
	if existing != nil {
		return fiber.NewError(400, "username already taken")
	}

	created := time.Now()
	user := &database.User{
		ID:        database.CreateID("usr"),
		Name:      username,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		CreatedAt: created,
	}

	if err := star.Database.CreateUser(ctx, user); err != nil {
		return err
	}

	email := &database.UserEmail{
		ID:        database.CreateID("eml"),
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
		ID:          database.CreateID("acc"),
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
	url := star.Config.ServerURL + "/verify?email=" + request.Email + "&code=" + email.Code
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

	return ctx.JSON(fiber.Map{"success": true})
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
		ID:          database.CreateID("acc"),
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

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	accounts := star.Database.GetAccountsByUserID(ctx, user.ID)
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

func getUserEmails(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	emails := star.Database.GetEmailsByUserID(ctx, user.ID)
	return ctx.JSON(transformer.TransformEmails(emails))
}

func addUserEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var data types.AddUserEmailRequest
	if err := ctx.Bind().Body(&data); err != nil {
		return err
	}

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	// Prevent taking over an email that is either verified, or has recently been added within the last hour
	existing := star.Database.GetEmailByName(ctx, data.Email)
	canUse := existing == nil || !existing.Verified || existing.CreatedAt.Add(time.Hour).Before(time.Now())

	if !canUse {
		return fiber.NewError(400, "the provided email is either already in use or has recently been added to an account")
	}

	email := &database.UserEmail{
		ID:        database.CreateID("eml"),
		UserID:    user.ID,
		Email:     data.Email,
		Verified:  false,
		Primary:   false,
		Code:      email.CreateCode(star.Settings.EmailCodeLength),
		CreatedAt: time.Now(),
	}

	if err := star.Database.CreateEmail(ctx, email); err != nil {
		return err
	}

	// Send verification email
	url := star.Config.ServerURL + "/verify?email=" + data.Email + "&code=" + email.Code
	if err := star.Email.Send(
		data.Email,
		"Verify your email address",
		emails.VerificationEmail,
		map[string]string{
			"Name":            user.Name,
			"VerificationURL": url,
			"Code":            email.Code,
		},
	); err != nil {
		slog.Warn("failed to send verification email", "err", err, "email", data.Email, "code", email.Code)
	}

	return ctx.JSON(transformer.TransformEmail(email))
}

func deleteUserEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != user.ID {
		return fiber.NewError(401, "you are not permitted to delete this email")
	}

	if err := star.Database.DeleteEmail(ctx, emailID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}

func updatePrimaryEmail(ctx fiber.Ctx) error {
	star := common.GetState(ctx)
	emailID := ctx.Params("email")

	user, err := request.GetUser(ctx)
	if err != nil {
		return err
	}

	email := star.Database.GetEmailByID(ctx, emailID)
	if email == nil {
		return fiber.NewError(404, "email not found")
	}

	if email.UserID != user.ID {
		return fiber.NewError(401, "the provided email is not associated with your account")
	}

	if err := star.Database.SetPrimaryEmail(ctx, user.ID, emailID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"success": true,
	})
}
