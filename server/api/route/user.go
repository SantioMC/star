package route

import (
	"net/mail"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/server/api/helper"
	"github.com/tubblelive/star/server/state"
	"github.com/tubblelive/star/server/verifier"
)

func UserRouter(r fiber.Router) {
	r.Get("/me", getMe)
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

	if fingerprint := body.Fingerprint; fingerprint != nil {
		err = verifier.VerifyFingerprint(*fingerprint)
		if err != nil {
			return helper.NewHttpError(400, "Invalid fingerprint, "+err.Error())
		}
	}

	userStore := state.StarState.Database.UserStore
	existing, err := userStore.GetUserByEmail(email.Address)
	
	if err != nil {
		return err
	}

	if existing != nil {
		return helper.NewHttpError(409, "The email provided is already in use")
	}

	user, err := userStore.CreateUser(
		body.Username, email.Address, body.Password,
	)

	if err != nil {
		return err
	}

	ctx.JSON(fiber.Map{
		"username": body.Username,
		"email": email.Address,
		"id": user.ID,
	})

	return nil
}

func getMe(ctx *fiber.Ctx) error {
	id, err := helper.GetUserId(ctx)
	if err != nil {
		return err
	}

	user, err := state.StarState.Database.UserStore.GetUserByID(id)
	if err != nil {
		return err
	}

	ctx.JSON(user)
	return nil
}

type RegisterAccountRequest struct {
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	Password     string  `json:"password"`
	Fingerprint  *string `json:"fingerprint"`
}
