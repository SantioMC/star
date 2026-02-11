package star

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tubblelive/star/star/api"
	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database"
	"github.com/tubblelive/star/star/database/seeder"
	"github.com/tubblelive/star/star/service"
	"github.com/tubblelive/star/star/state"
)

func StartServer(config *config.Config) {
	dialector, protocol, err := database.CreateDialector(config.Database)
	if err != nil {
		slog.Error("failed to create dialector", "error", err)
		return
	}

	slog.Info("connecting to database", "protocol", protocol)

	db, firstStart, err := database.Connect(*dialector, config)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	state.Database = db
	if err := seeder.SeedDatabase(firstStart, config); err != nil {
		slog.Error("failed to seed database", "error", err)
		return
	}

	signingKey, err := service.GetSigningKey()
	if err != nil {
		slog.Error("failed to get signing key", "error", err)
		return
	}

	server := &state.Server{
		Config:     config,
		SigningKey: signingKey,
	}

	state.StarState = server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	stopTicker := database.StartTicker(db)
	go api.StartAPIServer(server)

	<-interrupt

	println("")
	slog.Info("star authorization server shutting down")

	stopTicker <- true
	sqlDb, err := state.Database.DB.DB()

	if err == nil {
		sqlDb.Close()
	}
}
