package route

import (
	"net/mail"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/database/model"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/verifier"
)

func UserRouter(r fiber.Router) {
	// Routes /api/user/me
	r.Get("/me", getMe)
	r.Patch("/me", patchMe)
	r.Post("/me/password", updatePassword)
	r.Delete("/me", deleteMe)

	// Routes /api/user/register
	r.Post("/register", limiter.New(helper.HeavyLimiter), registerAccount)

	// Routes privileged endpoints
	r.Get("/", getAllUsers)
	r.Get("/search", searchUsers)
	r.Get("/:user", getUser)
	r.Patch("/:user", patchUser)
	r.Delete("/:user", deleteUser)
	r.Post("/:user/password", updateUserPassword)
	r.Post("/:user/suspend", suspendUser)
	r.Post("/:user/unsuspend", unsuspendUser)
	r.Post("/:user/invalidate", invalidateUserSessions)
}

func registerAccount(ctx *fiber.Ctx) error {
	body := new(RegisterAccountRequest)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	captcha := state.StarState.Config.CaptchaVerifier
	if captcha != nil {
		token := ctx.Get("X-Turnstile-Token", "")
		valid := captcha.ValidateToken(token)

		if !valid {
			return helper.NewHttpError(400, "Invalid captcha token provided")
		}
	}

	email, err := mail.ParseAddress(body.Email)
	if err != nil {
		return helper.NewHttpError(400, "Invalid email address provided")
	}

	err = verifier.VerifyUsername(body.Username)
	if err != nil {
		return helper.NewHttpError(400, "Invalid username, "+err.Error())
	}

	err = verifier.VerifyPassword(body.Password)
	if err != nil {
		return helper.NewHttpError(400, "Invalid password, "+err.Error())
	}

	existing, err := service.GetUserByEmail(email.Address)

	if err != nil {
		return err
	}

	if existing != nil {
		return helper.NewHttpError(409, "The email provided is already in use")
	}

	user, _, err := service.CreateUser(
		body.Username, email.Address, body.Password,
	)

	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"username": body.Username,
		"email":    email.Address,
		"id":       user.ID,
	})
}

func getMe(ctx *fiber.Ctx) error {
	id, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	user, err := service.GetUserByID(id)
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(500, "failed to find user by id")
	}

	if !helper.HasPermission(ctx, scopes.ScopeReadEmail) {
		user.Email = ""
		user.EmailVerified = true
	}

	sessions := []SessionResponse{}
	if helper.HasPermission(ctx, scopes.ScopeReadSessions) {
		sessions = findSessions(user.ID)
	}

	ctx.JSON(UserData{
		User: *user,

		Metadata:       user.GetMetadata(true),
		Sessions:       sessions,
		ProfilePicture: user.GenerateProfilePicture(),
	})

	return nil
}

func patchMe(ctx *fiber.Ctx) error {
	id, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	user, err := service.GetUserByID(id)
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(500, "failed to find user by id")
	}

	var body UpdateUserRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Username != nil {
		existing, err := service.GetUserByUsername(*body.Username)
		if err != nil {
			return err
		}

		if existing != nil && existing.ID != user.ID {
			return helper.NewHttpError(400, "username is already in use")
		}

		if err := verifier.VerifyUsername(*body.Username); err != nil {
			return err
		}

		user.Username = *body.Username
	}

	if body.Picture != nil {
		if err := verifier.VerifyURL(*body.Picture); err != nil {
			return helper.NewHttpError(400, err.Error())
		}

		user.ProfilePicture = *body.Picture
	}

	if body.FirstName != nil && len(*body.FirstName) < 64 {
		user.FirstName = *body.FirstName
	}

	if body.LastName != nil && len(*body.LastName) < 64 {
		user.LastName = *body.LastName
	}

	if err := service.SaveUser(*user); err != nil {
		return err
	}

	return ctx.JSON(user)
}

func deleteMe(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeDeleteUser); err != nil {
		return err
	}

	//todo: send email confirmation
	return nil
}

func updatePassword(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeChangePassword); err != nil {
		return err
	}

	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	user, err := service.GetUserByID(userId)
	if err != nil {
		return err
	}

	var body ChangePasswordRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	success, err := service.VerifyPassword(user, body.CurrentPassword)
	if err != nil {
		return err
	}

	if !success {
		return helper.NewHttpError(403, "your current password does not match!")
	}

	if err := verifier.VerifyPassword(body.NewPassword); err != nil {
		return helper.NewHttpError(400, err.Error())
	}

	if err := service.SetPassword(user, body.NewPassword); err != nil {
		return err
	}

	if err := service.InvalidateAllSessions(userId); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func getAllUsers(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadUsers); err != nil {
		return err
	}

	page := ctx.Query("page", "1")
	pageNum, err := strconv.Atoi(page)

	users, err := service.GetAllUsers(pageNum, 25)
	if err != nil {
		return err
	}

	return ctx.JSON(toAdminUserResponse(users))
}

func getUser(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(404, "user not found")
	}

	adminUser := toAdminUserResponse([]model.User{*user})[0]
	return ctx.JSON(adminUser)
}

func updateUserPassword(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(404, "user not found")
	}

	var body AdminChangePasswordRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if len(strings.TrimSpace(body.NewPassword)) == 0 {
		return helper.NewHttpError(400, "password can not be an empty string")
	}

	if err := service.SetPassword(user, body.NewPassword); err != nil {
		return err
	}

	if err := service.InvalidateAllSessions(user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func patchUser(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(404, "user not found")
	}

	var body UpdateUserRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Username != nil {
		existing, err := service.GetUserByUsername(*body.Username)
		if err != nil {
			return err
		}

		if existing != nil && existing.ID != user.ID {
			return helper.NewHttpError(400, "username is already in use")
		}

		user.Username = *body.Username
	}

	if body.Picture != nil {
		if err := verifier.VerifyURL(*body.Picture); err != nil {
			return helper.NewHttpError(400, err.Error())
		}

		user.ProfilePicture = *body.Picture
	}

	if body.FirstName != nil {
		user.FirstName = *body.FirstName
	}

	if body.LastName != nil {
		user.LastName = *body.LastName
	}

	if err := service.SaveUser(*user); err != nil {
		return err
	}

	return ctx.JSON(user)
}

func deleteUser(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminDeleteUsers); err != nil {
		return err
	}

	userId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	target := strings.TrimSpace(ctx.Params("user"))

	if strings.EqualFold(userId, target) {
		return helper.NewHttpError(400, "you can not delete yourself")
	}

	if err := service.DeleteUser(target); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func suspendUser(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user.Suspension != nil {
		return helper.NewHttpError(400, "user is already suspended")
	}

	issuerId, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	if strings.EqualFold(user.ID, issuerId) {
		return helper.NewHttpError(400, "you can not suspend yourself")
	}

	var body SuspendUserRequest
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if err := service.SuspendUser(user.ID, issuerId, body.Reason, body.Duration); err != nil {
		return err
	}

	if err := service.InvalidateAllSessions(user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func unsuspendUser(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user.Suspension == nil {
		return helper.NewHttpError(400, "user is not suspended")
	}

	if err := service.UnsuspendUser(user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func searchUsers(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminReadUsers); err != nil {
		return err
	}

	query := ctx.Query("query", "")
	page := ctx.Query("page", "1")
	pageNum, err := strconv.Atoi(page)

	users, err := service.SearchUsers(query, pageNum, 25)
	if err != nil {
		return err
	}

	return ctx.JSON(toAdminUserResponse(users))
}

func invalidateUserSessions(ctx *fiber.Ctx) error {
	if err := helper.EnforcePermission(ctx, scopes.ScopeAdminModifyUsers); err != nil {
		return err
	}

	user, err := service.GetUserByID(ctx.Params("user"))
	if err != nil {
		return err
	}

	if user == nil {
		return helper.NewHttpError(404, "user not found")
	}

	if err := service.InvalidateAllSessions(user.ID); err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{"success": true})
}

func toAdminUserResponse(users []model.User) []UserAdminData {
	data := []UserAdminData{}

	for _, user := range users {
		sessions := []SessionResponse{}

		for _, session := range user.Sessions {
			sessions = append(sessions, toSessionResponse(&session))
		}

		data = append(data, UserAdminData{
			UserData: UserData{
				User:           user,
				Metadata:       user.GetMetadata(false),
				ProfilePicture: user.GenerateProfilePicture(),
				Sessions:       sessions,
			},
			LastLoginAt: user.LastLoginAt,
			Suspension:  user.Suspension,
		})
	}

	return data
}

type UserAdminData struct {
	UserData

	LastLoginAt int64             `json:"last_login_at"`
	Suspension  *model.Suspension `json:"suspension"`
}

type UserData struct {
	model.User

	Metadata       model.UserMetadata `json:"metadata"`
	ProfilePicture string             `json:"profile_picture"`
	Sessions       []SessionResponse  `json:"sessions"`
}

type RegisterAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AdminChangePasswordRequest struct {
	NewPassword string `json:"password"`
}

type SuspendUserRequest struct {
	Reason   string `json:"reason"`
	Duration *int64 `json:"duration"`
}

type UpdateUserRequest struct {
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Picture   *string `json:"profile_picture"`
}
