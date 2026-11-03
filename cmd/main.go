package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/cmd/initializer"
	"github.com/tubblelive/star/internal/server"
)

func main() {
	started := time.Now()
	star := initializer.PrepareServer()

	if star == nil {
		os.Exit(1)
	}

	app, err := server.CreateServer(star)
	if err != nil {
		slog.Error("failed to create server", "err", err)
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	app.Hooks().OnListen(func(data fiber.ListenData) error {
		elapsed := time.Since(started)
		slog.Info("star authorization server started", "startup", elapsed, "port", data.Port)
		return nil
	})

	go func() {
		address := fmt.Sprintf("[::]:%d", star.Config.Port)

		if err := app.Listen(address); err != nil {
			slog.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	<-shutdown
	slog.Info("shutting down star authorization server")

	if err := app.ShutdownWithTimeout(time.Second * 15); err != nil {
		slog.Error("failed to gracefully shutdown server", "err", err)
	}

	star.Database.Close()
}
