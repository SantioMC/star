package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

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
)

type LoginFlow struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Stage     LoginFlowStage `json:"stage"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
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

func (db *Database) DeleteLoginFlowByID(ctx context.Context, id string) error {
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM login_flows WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) CreateLoginFlow(ctx context.Context, loginFlow *LoginFlow) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO login_flows (
			id, user_id, stage, created_at
		) VALUES ($1, $2, $3, $4)
		`,
		loginFlow.ID,
		loginFlow.UserID,
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
