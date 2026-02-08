package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/tubblelive/star/star"
	"github.com/tubblelive/star/star/config"
)

func main() {
	setupLogger()
	slog.Info("starting star authorization server")

	config, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	star.StartServer(config)
}

func setupLogger() {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.TimeOnly,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
