package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tubblelive/star/config"
	"github.com/tubblelive/star/server/api"
	"github.com/tubblelive/star/server/database"
	"github.com/tubblelive/star/server/state"
)

func StartServer(config *config.Config) {
	dialector, protocol, err := database.CreateDialector(config.Database)
	if err != nil {
		slog.Error("failed to create dialector", "error", err)
		return
	}

	slog.Info("connecting to database", "protocol", protocol)

	db, err := database.Connect(*dialector, config)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	signingKey, err := db.KeyStore.GetSigningKey()
	if err != nil {
		slog.Error("failed to get signing key", "error", err)
		return
	}

	server := &state.Server{
		Config:   config,
		Database: db,
		SigningKey: signingKey,
	}

	state.StarState = server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go api.StartAPIServer(server)

	user, err := server.Database.UserStore.GetUserByEmail("root@star")
	if err != nil {
		slog.Error("failed to get super user", "err", err)
	} else {
		slog.Info("got super user", "name", user.Username)
	}

	<-interrupt

	println("")
	slog.Info("star authorization server shutting down")
	sqlDb, err := server.Database.DB.DB()

	if err == nil {
		sqlDb.Close()
	}
}
