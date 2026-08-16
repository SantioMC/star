package server

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/login"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/types"
)

func loginRouter(router fiber.Router) {
	router.Post("/", limiter.New(middleware.HeavyRateLimiter), middleware.RequireCaptcha, loginUser)
	router.Post("/totp", limiter.New(middleware.HeavyRateLimiter), loginFlowTotp)
}

func findUserByIdentifier(ctx fiber.Ctx, identifier string) *database.User {
	star := common.GetState(ctx)

	matches := star.Database.GetUsersByName(ctx, identifier)
	for _, user := range matches {
		if user.Name == identifier {
			return &user
		}
	}

	if len(matches) == 1 {
		return &matches[0]
	} else if len(matches) > 1 {
		slog.Warn("login attempt failed", "reason", "username matches multiple users", "identifier", identifier)
		return nil
	}

	if !strings.Contains(identifier, "@") {
		return nil
	}

	matches = star.Database.GetUsersByVerifiedEmail(ctx, identifier)
	if len(matches) == 1 {
		return &matches[0]
	} else if len(matches) > 1 {
		slog.Warn("login attempt failed", "reason", "email matches multiple users", "identifier", identifier)
	}

	return nil
}

func loginUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	user := findUserByIdentifier(ctx, request.Username)
	if user == nil {
		slog.Warn("login attempt failed", "reason", "unknown login identifier")
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

	// We have successfully passed first factor at this point
	factors, err := mfa.GetMFA(ctx, star, user.ID)
	if err != nil {
		slog.Error("login attempt failed", "reason", "critical mfa error has occurred", "err", err)
		return err
	}

	if len(factors) == 0 {
		// We have no MFA available, we have completed all factors
		if err := login.Preverify(ctx, star, user, account); err != nil {
			return err
		}

		// All done, allow them through
		flow, err := login.CreateSuccessfulStage(ctx, star, user, account)
		if err != nil {
			return err
		}

		return ctx.JSON(flow)
	}

	flow, err := login.CreateRequireMFAStage(ctx, star, user, account, factors)
	if err != nil {
		return err
	}

	return ctx.JSON(flow)
}

func loginFlowTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginFlowTotpRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	loginFlow := star.Database.GetLoginStageByToken(ctx, request.Flow)
	if loginFlow == nil || loginFlow.Stage == database.StageCompleted {
		return fiber.NewError(400, "invalid login flow identifier provided")
	}

	if loginFlow.Stage != database.StageMFARequired {
		return fiber.NewError(400, "login flow is not requesting a totp")
	}

	totp := star.Database.GetUserTotp(ctx, loginFlow.UserID)
	if totp == nil {
		return fiber.NewError(400, "totp is not a valid factor")
	}

	if !totp.Verified {
		slog.Error("failed to attempt totp verification, totp was not verified but still attempted login with it", "user", loginFlow.UserID, "totp", totp.ID)
		return fiber.NewError(500, "totp is not verified, unable to continue")
	}

	if err := mfa.VerifyTotp(totp.Secret, request.Code); err != nil {
		return err
	}

	user := star.Database.GetUserByID(ctx, loginFlow.UserID)
	if user == nil {
		return fiber.NewError(500, "failed to retrieve your user details from database")
	}

	account := star.Database.GetInternalAccountByUserID(ctx, loginFlow.UserID)
	if account == nil {
		return fiber.NewError(500, "failed to retrieve your account details from database")
	}

	// User has completed totp, let's run some last verifications
	if err := login.Preverify(ctx, star, user, account); err != nil {
		return err
	}

	// Complete the login
	stage, err := login.CreateSuccessfulStage(ctx, star, user, account)
	if err != nil {
		return err
	}

	// Mark the stage as now complete
	loginFlow.Stage = database.StageCompleted
	if err := star.Database.UpdateLoginFlow(ctx, loginFlow); err != nil {
		return err
	}

	return ctx.JSON(stage)
}
