package initializer

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/tubblelive/star/internal/cache"
	"github.com/tubblelive/star/internal/cleanup"
	"github.com/tubblelive/star/internal/config"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/database/seed"
	"github.com/tubblelive/star/internal/passkey"
	"github.com/tubblelive/star/internal/revocation"
	"github.com/tubblelive/star/internal/state"
)

func PrepareServer() *state.StarServer {
	tintHandler := tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.TimeOnly,
		Level:      slog.LevelInfo,
	})

	logger := slog.New(tintHandler)
	slog.SetDefault(logger)

	if err := godotenv.Load(); errors.Is(err, fs.ErrNotExist) {
		slog.Debug("no .env file found, using environment variables only")
	} else if err != nil {
		slog.Error("failed to load .env file", "err", err)
		return nil
	}

	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		return nil
	}

	db, err := database.Connect(context.Background(), config)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		return nil
	}

	if err := database.Migrate(db); err != nil {
		slog.Error("failed to migrate database", "err", err)
		return nil
	}

	relyingParty, err := passkey.NewRelyingParty(config)
	if err != nil {
		slog.Error("failed to configure passkeys", "err", err)
		return nil
	}

	star := state.StarServer{
		Database:    db,
		GeoService:  config.Geolocation.CreateService(),
		Captcha:     config.Captcha.CreateVerifier(),
		Email:       config.Email.CreateSender(),
		WebAuthn:    relyingParty,
		Config:      config,
		StartedAt:   time.Now(),
		Revocations: revocation.NewList(),
	}

	if err := seed.Seed(context.Background(), &star); err != nil {
		slog.Error("failed to seed database", "err", err)
		return nil
	}

	// Get our cache up to date, and cleanup dead and old records
	cache.StartTicker(&star)
	cleanup.StartTicker(&star)

	if star.DefaultApp.Load() == nil {
		slog.Error("failed to load the built-in star application")
		return nil
	}

	if star.Branding.Load() == nil {
		slog.Error("failed to load the server branding")
		return nil
	}

	return &star
}
