package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/migrations"
	"github.com/vinovest/sqlx"
)

var alphabet = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

type Database struct {
	config *config.Config
	pool   *pgxpool.Pool
	db     *sqlx.DB
}

func (db *Database) DB() *sqlx.DB {
	return db.db
}

func (db *Database) Close() {
	db.db.Close()
	db.pool.Close()
}

func Connect(ctx context.Context, config *config.Config) (*Database, error) {
	var err error

	cfg, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(timeout); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	nativeDb := sql.OpenDB(stdlib.GetPoolConnector(pool))
	sqlx := sqlx.NewDb(nativeDb, "pgx")

	return &Database{
		config: config,
		pool:   pool,
		db:     sqlx,
	}, nil
}

func Migrate(pool *Database) error {
	db := sql.OpenDB(stdlib.GetPoolConnector(pool.pool))
	defer db.Close()

	goose.SetBaseFS(migrations.Filesystem)
	goose.SetTableName("migrations")
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil && !errors.Is(err, goose.ErrAlreadyApplied) {
		return err
	}

	return nil
}

// CreateID creates a random ID of the given kind. This is not cryptographically secure.
func CreateID(kind string) string {
	b := make([]rune, 16)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return fmt.Sprintf("%s_%s", kind, string(b))
}
