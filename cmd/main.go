package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tubblelive/star/cmd/initializer"
	"github.com/tubblelive/star/internal/server"
)

func main() {
	started := time.Now()
	star := initializer.PrepareServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	if star == nil {
		return
	}

	go func() {
		if err := server.StartServer(star); err != nil {
			slog.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	elapsed := time.Since(started)
	slog.Info("star authorization server started", "startup", elapsed)

	<-shutdown
}
