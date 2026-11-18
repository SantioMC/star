package login

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/mfa"
	"github.com/tubblelive/star/internal/state"
)

var ErrLoginFlowExpired = fiber.NewError(400, "Your sign in attempt expired, please sign in again")

type RequireMFAStage struct {
	Next    string          `json:"next"`
	Flow    string          `json:"flow"`
	Factors []mfa.MFAMethod `json:"factors"`
}

type MFAAttempt struct {
	Flow int
	User *database.LoginAttempts
}

func CreateRequireMFAStage(
	ctx fiber.Ctx,
	star *state.StarServer,
	user *database.User,
	account *database.Account,
	mfa []mfa.MFAMethod,
) (*RequireMFAStage, error) {
	token, err := createFlow(ctx, star, user, account, database.StageMFARequired)
	if err != nil {
		return nil, err
	}

	return &RequireMFAStage{
		Next:    "mfa_required",
		Flow:    token,
		Factors: mfa,
	}, nil
}

func CheckLockout(ctx fiber.Ctx, star *state.StarServer, userID string) error {
	attempts, err := star.Database.GetLoginAttempts(ctx, userID)
	if err != nil {
		return err
	} else if attempts != nil && attempts.IsLocked() {
		return ErrLoginLocked
	}

	return nil
}

func ReserveMFAAttempt(ctx fiber.Ctx, star *state.StarServer, loginFlow *database.LoginFlow) (*MFAAttempt, error) {
	userAttempt, err := ReserveAttempt(ctx, star, loginFlow.UserID)
	if err != nil {
		return nil, err
	}

	flowAttempt, reserved, err := star.Database.ReserveLoginFlowAttempt(ctx, loginFlow.ID)
	if err != nil {
		return nil, err
	} else if !reserved {
		return nil, ErrLoginFlowExpired
	}

	return &MFAAttempt{Flow: flowAttempt, User: userAttempt}, nil
}

func RecordMFAFailure(ctx fiber.Ctx, star *state.StarServer, loginFlow *database.LoginFlow, attempt *MFAAttempt) error {
	if attempt.Flow >= 5 {
		if _, err := star.Database.DeleteLoginFlowByID(ctx, loginFlow.ID); err != nil {
			return err
		}
	}

	ReportFailure(ctx, star, attempt.User, database.SecurityAlertMFABruteForce)
	return nil
}
