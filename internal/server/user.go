package server

import (
	"errors"
	"log/slog"
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

	existing := star.Database.GetUserByName(ctx, username)
	if existing != nil {
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

	users := star.Database.GetUsers(ctx, page, limit)
	return ctx.JSON(users)
}
