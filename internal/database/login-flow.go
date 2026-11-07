package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

const LoginFlowKey = "login"

type LoginFlowStage string

const (
	// Used to indicate that they are logged in and were provided a token - not to be stored in db
	StageCompleted LoginFlowStage = "complete"
	// Set when the user is required to enter a second factor for login
	StageMFARequired LoginFlowStage = "mfa_required"
	// Set when the user is required to change their password before being authenticated
	StagePasswordChange LoginFlowStage = "pass_change"
	StageBirthday       LoginFlowStage = "birthday"
	// Set when the user has signed in through an external provider and is waiting for the frontend to claim the flow
	StageProvider LoginFlowStage = "provider"
)

type LoginFlow struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id" db:"user_id"`
	AccountID *string        `json:"account_id" db:"account_id"`
	Token     string         `json:"token"`
	Stage     LoginFlowStage `json:"stage"`
	Attempts  int            `json:"attempts"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

func CreateLoginFlowToken() string {
	return cryptography.CreateSecureString(48)
}

func (db *Database) GetLoginStageByID(ctx context.Context, id string) *LoginFlow {
	loginFlow, err := sqlx.OneContext[LoginFlow](
		ctx,
		db.DB(),
		`SELECT * FROM login_flows WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get login flow by id", "err", err)
		return nil
	}

	return &loginFlow
}

func (db *Database) GetLoginStageByToken(ctx context.Context, token string) *LoginFlow {
	hashed := cryptography.HMAC(token, db.config.AppToken)

	loginFlow, err := sqlx.OneContext[LoginFlow](
		ctx,
		db.DB(),
		`SELECT * FROM login_flows WHERE token = $1 LIMIT 1`,
		hashed,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get login flow by token", "err", err)
		return nil
	}

	return &loginFlow
}

func (db *Database) DeleteLoginFlowByID(ctx context.Context, id string) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM login_flows WHERE id = $1`,
		id,
	)

	if err != nil {
		return false, err
	}

	deleted, err := result.RowsAffected()
	return deleted > 0, err
}

func (db *Database) ReserveLoginFlowAttempt(ctx context.Context, id string) (int, bool, error) {
	attempts, err := sqlx.OneContext[int](
		ctx,
		db.DB(),
		`UPDATE login_flows SET attempts = attempts + 1 WHERE id = $1 AND attempts < 5 RETURNING attempts`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	return attempts, true, nil
}

func (db *Database) CreateLoginFlow(ctx context.Context, loginFlow *LoginFlow) error {
	hashed := cryptography.HMAC(loginFlow.Token, db.config.AppToken)

	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO login_flows (
			id, user_id, account_id, token, stage, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		`,
		loginFlow.ID,
		loginFlow.UserID,
		loginFlow.AccountID,
		hashed,
		loginFlow.Stage,
		loginFlow.CreatedAt,
	)

	return err
}

func (db *Database) UpdateLoginFlow(ctx context.Context, loginFlow *LoginFlow) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE login_flows
		SET stage = $2
		WHERE id = $1
		`,
		loginFlow.ID,
		loginFlow.Stage,
	)

	return err
}

func (db *Database) CompleteRequiredPasswordChange(ctx context.Context, loginFlow *LoginFlow, hashed string, alg cryptography.Algorithm) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	if consumed, err := consumeLoginFlow(ctx, tx, loginFlow, StagePasswordChange); err != nil || !consumed {
		return false, err
	}

	updated, err := tx.ExecContext(
		ctx,
		`
		UPDATE accounts
		SET password = $2, password_alg = $3, password_change_required = FALSE
		WHERE user_id = $1 AND type = 'internal' AND password_change_required
		`,
		loginFlow.UserID,
		hashed,
		alg,
	)

	if err != nil {
		return false, err
	}

	if affected, err := updated.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	return true, tx.Commit()
}

func (db *Database) CompleteRequiredBirthday(ctx context.Context, loginFlow *LoginFlow, birthday *models.Date) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	if consumed, err := consumeLoginFlow(ctx, tx, loginFlow, StageBirthday); err != nil || !consumed {
		return false, err
	}

	updated, err := tx.ExecContext(
		ctx,
		`UPDATE users SET birthday = $2 WHERE id = $1 AND birthday IS NULL`,
		loginFlow.UserID,
		birthday,
	)

	if err != nil {
		return false, err
	}

	if affected, err := updated.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	return true, tx.Commit()
}

func consumeLoginFlow(ctx context.Context, execer sqlx.ExecerContext, loginFlow *LoginFlow, stage LoginFlowStage) (bool, error) {
	consumed, err := execer.ExecContext(
		ctx,
		`DELETE FROM login_flows WHERE id = $1 AND stage = $2`,
		loginFlow.ID,
		stage,
	)

	if err != nil {
		return false, err
	}

	affected, err := consumed.RowsAffected()
	return affected > 0, err
}

func (db *Database) DeleteDeadLoginFlows(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM login_flows WHERE created_at < $1`,
		time.Now().Add(-time.Hour),
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *Database) GetLoginFlowCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM login_flows`)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *Database) GetLoginFlowsByUserID(ctx context.Context, userID string) []LoginFlow {
	loginFlows, err := list[LoginFlow](
		ctx,
		db.DB(),
		`SELECT * FROM login_flows WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)

	if err != nil {
		slog.Error("failed to get user login flows", "err", err)
		return []LoginFlow{}
	}

	return loginFlows
}
