package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server"
)

func main() {
	setupLogger()
	slog.Info("starting star authorization server")

	config, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	server.StartServer(config)
}

func setupLogger() {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.TimeOnly,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}