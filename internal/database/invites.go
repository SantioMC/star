package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/cryptography"
	"github.com/vinovest/sqlx"
)

const UserInviteKey = "inv"

type UserInvite struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	InvitedBy     *string   `json:"invited_by" db:"invited_by"`
	InvitedByName *string   `json:"invited_by_name" db:"invited_by_name"`
	Token         string    `json:"-"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

const selectUserInvites = `
	SELECT invites.*, inviter.name AS invited_by_name
	FROM user_invites invites
	LEFT JOIN users inviter ON inviter.id = invites.invited_by
`

func CreateUserInviteToken() string {
	return cryptography.CreateSecureString(48)
}

func (db *Database) GetUserInvite(ctx context.Context, userID string) (*UserInvite, error) {
	invite, err := sqlx.OneContext[UserInvite](
		ctx,
		db.DB(),
		selectUserInvites+` WHERE invites.user_id = $1 LIMIT 1`,
		userID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &invite, nil
}

func (db *Database) GetUserInviteByToken(ctx context.Context, token string) *UserInvite {
	invite, err := sqlx.OneContext[UserInvite](
		ctx,
		db.DB(),
		selectUserInvites+` WHERE invites.token = $1 AND invites.expires_at > $2 LIMIT 1`,
		cryptography.HMAC(token, db.config.AppToken),
		time.Now(),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get user invite by token", "err", err)
		return nil
	}

	return &invite
}

func (db *Database) CreateInvitedUser(ctx context.Context, user *User, email *UserEmail, invite *UserInvite) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := db.insertUser(ctx, tx, user); err != nil {
		return duplicate(err)
	}

	if err := db.insertEmail(ctx, tx, email); err != nil {
		return duplicate(err)
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO user_invites (id, user_id, invited_by, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		invite.ID,
		invite.UserID,
		invite.InvitedBy,
		cryptography.HMAC(invite.Token, db.config.AppToken),
		invite.ExpiresAt,
		invite.CreatedAt,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) RegenerateUserInvite(ctx context.Context, invite *UserInvite) (bool, error) {
	result, err := db.DB().ExecContext(
		ctx,
		`UPDATE user_invites SET token = $2, expires_at = $3, invited_by = $4 WHERE user_id = $1`,
		invite.UserID,
		cryptography.HMAC(invite.Token, db.config.AppToken),
		invite.ExpiresAt,
		invite.InvitedBy,
	)

	if err != nil {
		return false, err
	}

	updated, err := result.RowsAffected()
	return updated > 0, err
}

func (db *Database) AcceptUserInvite(ctx context.Context, invite *UserInvite, user *User, account *Account) (bool, error) {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	consumed, err := tx.ExecContext(
		ctx,
		`DELETE FROM user_invites WHERE id = $1 AND expires_at > $2`,
		invite.ID,
		time.Now(),
	)

	if err != nil {
		return false, err
	}

	if affected, err := consumed.RowsAffected(); err != nil || affected == 0 {
		return false, err
	}

	if err := db.updateUser(ctx, tx, user); err != nil {
		return false, err
	}

	if err := db.insertAccount(ctx, tx, account); err != nil {
		return false, duplicate(err)
	}

	return true, tx.Commit()
}

func (db *Database) DeleteExpiredInvitedUsers(ctx context.Context) ([]User, error) {
	return list[User](
		ctx,
		db.DB(),
		`DELETE FROM users WHERE id IN (SELECT user_id FROM user_invites WHERE expires_at < $1) RETURNING *`,
		time.Now(),
	)
}
