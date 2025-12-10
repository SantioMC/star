package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/tubblelive/star/server/api/helper"
	"github.com/tubblelive/star/server/database/model"
	"github.com/tubblelive/star/server/state"
)

func AppRouter(r fiber.Router) {
	r.Post("/:app/login", limiter.New(helper.HeavyLimiter), login)
}

func login(ctx *fiber.Ctx) error {
	body := new(LoginRequest)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	app, err := getAppFromCtx(ctx)
	if err != nil {
		return helper.NewHttpError(404, "Invalid application slug provided")
	}

	userStore := state.StarState.Database.UserStore
	user, err := userStore.ValidateLogin(body.Username, body.Password)
	if err != nil {
		return helper.NewHttpError(401, "Invalid username or password")
	}

	refreshToken, accessToken, expiresAt, err := userStore.CreateRefreshToken(user, app)
	if err != nil {
		return err
	}

	ctx.JSON(LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken.Token,
		ExpiresAt:     expiresAt.Unix(),
		ApplicationID: app.ID,
		UserID:        user.ID,
	})

	return nil
}

func getAppFromCtx(ctx *fiber.Ctx) (*model.Application, error) {
	return state.StarState.Database.AppStore.GetAppBySlug(ctx.Params("app", ""))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	RefreshToken  string `json:"refresh_token"`
	AccessToken   string `json:"access_token"`
	ExpiresAt     int64  `json:"expires_at"`
	ApplicationID string `json:"app_id"`
	UserID        string `json:"user_id"`
}
