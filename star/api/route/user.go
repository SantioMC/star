package route

import (
	"net/mail"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/star/api/helper"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state"
	"github.com/tubblelive/star/star/state/scopes"
	"github.com/tubblelive/star/star/verifier"
)

func UserRouter(r fiber.Router) {
	r.Get("/me", getMe)
	r.Patch("/me", patchMe)
	r.Delete("/me", deleteMe)

	r.Post("/register", limiter.New(helper.HeavyLimiter), registerAccount)
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

	ctx.JSON(fiber.Map{
		"username": body.Username,
		"email":    email.Address,
		"id":       user.ID,
	})

	return nil
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

	ctx.JSON(user)
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

		if existing != nil {
			return helper.NewHttpError(400, "username is already in use")
		}

		user.Username = *body.Username
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

	ctx.JSON(user)
	return nil
}

func deleteMe(ctx *fiber.Ctx) error {
	helper.EnforcePermission(ctx, scopes.ScopeDeleteUser)

	//todo: send email confirmation
	return nil
}

type RegisterAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Username  *string `json:"username"`
	FirstName *string `json:"given_name"`
	LastName  *string `json:"family_name"`
}
