package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/tubblelive/star/internal/database/models"
	"github.com/vinovest/sqlx"
)

const UserKey = "usr"

type User struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	FirstName      *string      `json:"first_name" db:"first_name"`
	LastName       *string      `json:"last_name" db:"last_name"`
	ProfilePicture *string      `json:"profile_picture" db:"profile_picture"`
	Birthday       *models.Date `json:"birthday" db:"birthday"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

func (db *Database) GetUserByID(ctx context.Context, id string) *User {
	user, err := sqlx.OneContext[User](
		ctx,
		db.DB(),
		`SELECT * FROM users WHERE id = $1 LIMIT 1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get user by id", "err", err)
		return nil
	}

	return &user
}

func (db *Database) GetUsers(ctx context.Context, page int, limit int) []User {
	users, err := list[User](
		ctx,
		db.DB(),
		`SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, (page-1)*limit,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []User{}
	} else if err != nil {
		slog.Error("failed to get users", "err", err)
		return []User{}
	}

	return users
}

func (db *Database) SearchUsers(ctx context.Context, query string, page int, limit int) []User {
	users, err := list[User](
		ctx,
		db.DB(),
		`
		SELECT * FROM users
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
		`,
		query, limit, (page-1)*limit,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []User{}
	} else if err != nil {
		slog.Error("failed to search users", "err", err)
		return []User{}
	}

	return users
}

func (db *Database) GetUserByName(ctx context.Context, name string) *User {
	user, err := sqlx.OneContext[User](
		ctx,
		db.DB(),
		`SELECT * FROM users WHERE LOWER(name) = LOWER($1) LIMIT 1`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		slog.Error("failed to get user by name", "err", err)
		return nil
	}

	return &user
}

func (db *Database) GetUsersByName(ctx context.Context, name string) []User {
	users, err := list[User](
		ctx,
		db.DB(),
		`SELECT * FROM users WHERE LOWER(name) = LOWER($1)`,
		name,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []User{}
	} else if err != nil {
		slog.Error("failed to get users by name", "err", err)
		return []User{}
	}

	return users
}

func (db *Database) GetUsersByVerifiedEmail(ctx context.Context, email string) []User {
	users, err := list[User](
		ctx,
		db.DB(),
		`
		SELECT DISTINCT users.* FROM users
		JOIN user_emails ON user_emails.user_id = users.id
		WHERE LOWER(user_emails.email) = LOWER($1) AND user_emails.verified = true
		`,
		email,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return []User{}
	} else if err != nil {
		slog.Error("failed to get users by email", "err", err)
		return []User{}
	}

	return users
}

func (db *Database) DeleteUser(ctx context.Context, id string) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := db.revokeUser(ctx, tx, id, RevocationReasonUserDeleted); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *Database) UpdateUser(ctx context.Context, user *User) error {
	return db.updateUser(ctx, db.DB(), user)
}

func (db *Database) updateUser(ctx context.Context, execer sqlx.ExecerContext, user *User) error {
	_, err := execer.ExecContext(
		ctx,
		`
		UPDATE users
		SET name = $2, first_name = $3, last_name = $4, profile_picture = $5, birthday = $6
		WHERE id = $1
		`,
		user.ID,
		user.Name,
		user.FirstName,
		user.LastName,
		user.ProfilePicture,
		user.Birthday,
	)

	return err
}

func (db *Database) CreateUserIdentity(ctx context.Context, user *User, email *UserEmail, account *Account) error {
	tx, err := db.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := db.insertUser(ctx, tx, user); err != nil {
		return duplicate(err)
	}

	if email != nil {
		if err := db.insertEmail(ctx, tx, email); err != nil {
			return duplicate(err)
		}
	}

	if err := db.insertAccount(ctx, tx, account); err != nil {
		return duplicate(err)
	}

	return tx.Commit()
}

func (db *Database) insertUser(ctx context.Context, execer sqlx.ExecerContext, user *User) error {
	_, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO users (id, name, first_name, last_name, profile_picture, birthday, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		user.ID,
		user.Name,
		user.FirstName,
		user.LastName,
		user.ProfilePicture,
		user.Birthday,
		user.CreatedAt,
	)

	return err
}

func (db *Database) GetUserCount(ctx context.Context) (int, error) {
	count, err := sqlx.OneContext[int](ctx, db.DB(), `SELECT COUNT(*) FROM users`)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *Database) FindUserIDs(ctx context.Context, search string, limit int) ([]string, error) {
	return list[string](
		ctx,
		db.DB(),
		`
		SELECT id FROM users
		WHERE name ILIKE '%' || $1 || '%'
		   OR id IN (SELECT user_id FROM user_emails WHERE LOWER(email) = LOWER($1))
		ORDER BY LOWER(name) = LOWER($1) DESC, LENGTH(name)
		LIMIT $2
		`,
		search,
		limit,
	)
}
