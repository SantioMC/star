package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tubblelive/star/internal/cryptography"
	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

const PasskeyChallengeKey = "pkc"

type PasskeyChallengeKind string

const (
	PasskeyChallengeRegister PasskeyChallengeKind = "register"
	PasskeyChallengeLogin    PasskeyChallengeKind = "login"
	PasskeyChallengeMFA      PasskeyChallengeKind = "mfa"
)

type PasskeyChallenge struct {
	ID          string                                `json:"id"`
	UserID      *string                               `json:"user_id" db:"user_id"`
	LoginFlowID *string                               `json:"login_flow_id" db:"login_flow_id"`
	Token       string                                `json:"token"`
	Kind        PasskeyChallengeKind                  `json:"kind"`
	Session     models.JSONData[webauthn.SessionData] `json:"session"`
	CreatedAt   time.Time                             `json:"created_at" db:"created_at"`
}

func CreatePasskeyChallengeToken() string {
	return cryptography.CreateSecureString(48)
}

func (db *Database) CreatePasskeyChallenge(ctx context.Context, challenge *PasskeyChallenge) error {
	hashed := cryptography.HMAC(challenge.Token, db.config.AppToken)

	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO passkey_challenges (
			id, user_id, login_flow_id, token, kind, session, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		challenge.ID,
		challenge.UserID,
		challenge.LoginFlowID,
		hashed,
		challenge.Kind,
		challenge.Session,
		challenge.CreatedAt,
	)

	return err
}

func (db *Database) ConsumePasskeyChallenge(ctx context.Context, token string, kind PasskeyChallengeKind) (*PasskeyChallenge, error) {
	hashed := cryptography.HMAC(token, db.config.AppToken)

	challenge, err := sqlx.OneContext[PasskeyChallenge](
		ctx,
		db.DB(),
		`
		DELETE FROM passkey_challenges
		WHERE token = $1 AND kind = $2 AND created_at > $3
		RETURNING *
		`,
		hashed,
		kind,
		time.Now().Add(-time.Minute*5),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (db *Database) DeleteDeadPasskeyChallenges(ctx context.Context) (int64, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM passkey_challenges WHERE created_at < $1`,
		time.Now().Add(-time.Minute*5),
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
