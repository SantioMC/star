package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

const UserKey = "usr"

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	FirstName      *string   `json:"first_name" db:"first_name"`
	LastName       *string   `json:"last_name" db:"last_name"`
	ProfilePicture *string   `json:"profile_picture" db:"profile_picture"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
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
	users, err := sqlx.ListContext[User](
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
	users, err := sqlx.ListContext[User](
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
		`SELECT * FROM users WHERE name = $1 LIMIT 1`,
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
	users, err := sqlx.ListContext[User](
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
	users, err := sqlx.ListContext[User](
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
	_, err := db.DB().ExecContext(
		ctx,
		`DELETE FROM users WHERE id = $1`,
		id,
	)

	return err
}

func (db *Database) UpdateUser(ctx context.Context, user *User) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		UPDATE users
		SET name = $2, first_name = $3, last_name = $4, profile_picture = $5
		WHERE id = $1
		`,
		user.ID,
		user.Name,
		user.FirstName,
		user.LastName,
		user.ProfilePicture,
	)

	return err
}

func (db *Database) CreateUser(ctx context.Context, user *User) error {
	_, err := db.DB().ExecContext(
		ctx,
		`
		INSERT INTO users (id, name, first_name, last_name, profile_picture, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		user.ID,
		user.Name,
		user.FirstName,
		user.LastName,
		user.ProfilePicture,
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
