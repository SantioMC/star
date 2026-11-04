package server

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/login"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/passkey"
	"github.com/tubblelive/star/internal/password"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/middleware"
	"github.com/tubblelive/star/internal/server/types"
	"github.com/tubblelive/star/internal/state"
)

func loginRouter(router fiber.Router) {
	router.Post("/", middleware.RateLimit(10), middleware.RequireCaptcha, loginUser)
	router.Post("/totp", middleware.RateLimit(10), loginFlowTotp)
	router.Post("/backup", middleware.RateLimit(10), loginFlowBackupCode)
	router.Post("/password", middleware.RateLimit(10), loginFlowPassword)
	router.Post("/passkey/options", middleware.RateLimit(10), passkeyLoginOptions)
	router.Post("/passkey", middleware.RateLimit(10), loginPasskey)

	router.Get("/providers", getLoginProviders)
	router.Post("/provider/complete", middleware.RateLimit(10), completeProviderLogin)
	router.Get("/provider/:provider", middleware.RateLimit(10), startProviderLogin)
	router.Get("/provider/:provider/callback", providerLoginCallback)
	router.Post("/provider/:provider/callback", providerLoginCallback)
}

func findUserByIdentifier(ctx fiber.Ctx, identifier string) *database.User {
	star := common.GetState(ctx)

	if !strings.Contains(identifier, "@") {
		return star.Database.GetUserByName(ctx, identifier)
	}

	matches := star.Database.GetUsersByVerifiedEmail(ctx, identifier)
	if len(matches) != 1 {
		return nil
	}

	return &matches[0]
}

func loginUser(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	user := findUserByIdentifier(ctx, request.Username)
	if user == nil {
		cryptography.SimulateVerifyHash(request.Password)
		slog.Warn("login attempt failed", "reason", "unknown login identifier")
		return fiber.NewError(401, "Invalid username or password")
	}

	account := star.Database.GetInternalAccountByUserID(ctx, user.ID)
	if account == nil || account.Password == nil {
		cryptography.SimulateVerifyHash(request.Password)
		slog.Warn("login attempt failed", "reason", "no internal account")
		return fiber.NewError(401, "Invalid username or password")
	}

	attempt, err := login.ReserveAttempt(ctx, star, user.ID)
	if err != nil {
		return err
	}

	success := cryptography.VerifyHash(request.Password, *account.Password)
	if !success {
		slog.Warn("login attempt failed", "reason", "invalid password")
		login.ReportFailure(ctx, star, user.ID, attempt, database.SecurityAlertPasswordBruteForce)
		return fiber.NewError(401, "Invalid username or password")
	}

	if err := star.Database.ReleaseLoginAttempt(ctx, user.ID); err != nil {
		return err
	}

	// We have successfully passed first factor at this point
	return continueLogin(ctx, star, user, account)
}

func continueLogin(ctx fiber.Ctx, star *state.StarServer, user *database.User, account *database.Account) error {
	factors, err := mfa.GetMFA(ctx, star, user.ID)
	if err != nil {
		slog.Error("login attempt failed", "reason", "critical mfa error has occurred", "err", err)
		return err
	}

	if len(factors) == 0 {
		// We have no MFA available, we have completed all factors
		stage, err := finishLogin(ctx, star, user, account)
		if err != nil {
			return err
		}

		return ctx.JSON(stage)
	}

	flow, err := login.CreateRequireMFAStage(ctx, star, user, account, factors)
	if err != nil {
		return err
	}

	return ctx.JSON(flow)
}

func loginFlowTotp(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginFlowCodeRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	loginFlow, err := pendingFlow(ctx, star, request.Flow, database.StageMFARequired)
	if err != nil {
		return err
	}

	totp := star.Database.GetUserTotp(ctx, loginFlow.UserID)
	if totp == nil {
		return fiber.NewError(400, "totp is not a valid factor")
	}

	if !totp.Verified {
		slog.Error("failed to attempt totp verification, totp was not verified but still attempted login with it", "user", loginFlow.UserID, "totp", totp.ID)
		return fiber.NewError(500, "totp is not verified, unable to continue")
	}

	attempt, err := login.ReserveMFAAttempt(ctx, star, loginFlow)
	if err != nil {
		return err
	}

	if err := mfa.VerifyTotp(ctx, star, totp, request.Code); errors.Is(err, mfa.ErrInvalidTotp) {
		return failMFAFlow(ctx, star, loginFlow, attempt, "Invalid authentication code, please try again")
	} else if err != nil {
		return err
	}

	return completeMFAFlow(ctx, star, loginFlow)
}

func loginFlowBackupCode(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginFlowCodeRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	loginFlow, err := pendingFlow(ctx, star, request.Flow, database.StageMFARequired)
	if err != nil {
		return err
	}

	attempt, err := login.ReserveMFAAttempt(ctx, star, loginFlow)
	if err != nil {
		return err
	}

	used, err := mfa.UseBackupCode(ctx, star, loginFlow.UserID, request.Code)
	if err != nil {
		return err
	} else if !used {
		return failMFAFlow(ctx, star, loginFlow, attempt, "Invalid backup code, please try again")
	}

	slog.Info("backup code used to sign in", "user", loginFlow.UserID)
	return completeMFAFlow(ctx, star, loginFlow)
}

func loginFlowPassword(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.LoginFlowPasswordRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	loginFlow, err := pendingFlow(ctx, star, request.Flow, database.StagePasswordChange)
	if err != nil {
		return err
	}

	account := star.Database.GetInternalAccountByUserID(ctx, loginFlow.UserID)
	if account == nil || account.Password == nil {
		return login.ErrLoginFlowExpired
	}

	sanitized, err := password.ValidatePassword(request.Password, star.Settings.Load().GetPasswordSettings())
	if err != nil {
		return fiber.NewError(400, err.Error())
	}

	if cryptography.VerifyHash(sanitized, *account.Password) {
		return fiber.NewError(400, "Your new password must be different from your current password")
	}

	hashed, alg := cryptography.HashString(sanitized)

	completed, err := star.Database.CompleteRequiredPasswordChange(ctx, loginFlow, hashed, alg)
	if err != nil {
		return err
	} else if !completed {
		return login.ErrLoginFlowExpired
	}

	ipAddress := star.ClientIP(ctx)
	slog.Info("required password change completed", "user", loginFlow.UserID, "ip_address", ipAddress)
	notifyPasswordChanged(ctx, star, loginFlow.UserID, ipAddress)

	stage, err := completeUserLogin(ctx, star, loginFlow.UserID, loginFlow.AccountID)
	if err != nil {
		return err
	}

	return ctx.JSON(stage)
}

func pendingFlow(ctx fiber.Ctx, star *state.StarServer, token string, stage database.LoginFlowStage) (*database.LoginFlow, error) {
	loginFlow := star.Database.GetLoginStageByToken(ctx, token)
	if loginFlow == nil || loginFlow.Stage == database.StageCompleted {
		return nil, fiber.NewError(400, "invalid login flow identifier provided")
	}

	if loginFlow.Stage != stage {
		return nil, fiber.NewError(400, "login flow is not waiting on this step")
	}

	if loginFlow.Attempts >= 5 || time.Since(loginFlow.CreatedAt) > time.Minute*10 {
		return nil, login.ErrLoginFlowExpired
	}

	if err := login.CheckLockout(ctx, star, loginFlow.UserID); err != nil {
		return nil, err
	}

	return loginFlow, nil
}

func failMFAFlow(ctx fiber.Ctx, star *state.StarServer, loginFlow *database.LoginFlow, attempt *login.MFAAttempt, message string) error {
	if err := login.RecordMFAFailure(ctx, star, loginFlow, attempt); err != nil {
		return err
	}

	return fiber.NewError(401, message)
}

func passkeyLoginOptions(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.PasskeyOptionsRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if request.Flow == "" {
		challenge, options, err := passkey.BeginLogin(ctx, star)
		if err != nil {
			return err
		}

		return ctx.JSON(fiber.Map{
			"challenge": challenge,
			"options":   options,
		})
	}

	loginFlow, err := pendingFlow(ctx, star, request.Flow, database.StageMFARequired)
	if err != nil {
		return err
	}

	challenge, options, err := passkey.BeginMFA(ctx, star, loginFlow)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"challenge": challenge,
		"options":   options,
	})
}

func loginPasskey(ctx fiber.Ctx) error {
	star := common.GetState(ctx)

	var request types.PasskeyLoginRequest
	if err := ctx.Bind().Body(&request); err != nil {
		return err
	}

	if request.Flow == "" {
		userID, err := passkey.FinishLogin(ctx, star, request.Challenge, request.Credential)
		if err != nil {
			return err
		}

		stage, err := completeUserLogin(ctx, star, userID, nil)
		if err != nil {
			return err
		}

		return ctx.JSON(stage)
	}

	loginFlow, err := pendingFlow(ctx, star, request.Flow, database.StageMFARequired)
	if err != nil {
		return err
	}

	if err := passkey.FinishMFA(ctx, star, request.Challenge, loginFlow, request.Credential); err != nil {
		return err
	}

	return completeMFAFlow(ctx, star, loginFlow)
}

func completeUserLogin(ctx fiber.Ctx, star *state.StarServer, userID string, accountID *string) (any, error) {
	user := star.Database.GetUserByID(ctx, userID)
	if user == nil {
		return nil, fiber.NewError(500, "failed to retrieve your user details from database")
	}

	var account *database.Account
	if accountID != nil {
		account = star.Database.GetAccountByID(ctx, *accountID)
	} else {
		account = star.Database.GetInternalAccountByUserID(ctx, userID)
	}

	return finishLogin(ctx, star, user, account)
}

func finishLogin(ctx fiber.Ctx, star *state.StarServer, user *database.User, account *database.Account) (any, error) {
	// User has completed every factor, let's run some last verifications
	if err := login.Preverify(ctx, star, user, account); err != nil {
		return nil, err
	}

	internal := star.Database.GetInternalAccountByUserID(ctx, user.ID)
	if internal != nil && internal.PasswordChangeRequired {
		return login.CreateRequirePasswordChangeStage(ctx, star, user, account)
	}

	stage, err := login.CreateSuccessfulStage(ctx, star, user, account)
	if err != nil {
		return nil, err
	}

	setSessionCookie(ctx, star, stage.Session)
	return stage, nil
}

func completeMFAFlow(ctx fiber.Ctx, star *state.StarServer, loginFlow *database.LoginFlow) error {
	stage, err := completeUserLogin(ctx, star, loginFlow.UserID, loginFlow.AccountID)
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
