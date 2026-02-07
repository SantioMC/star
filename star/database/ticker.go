package database

import (
	"log/slog"
	"time"
)

func StartTicker(db *Database) chan bool {
	ticker := time.NewTicker(5 * time.Minute)
	shutdown := make(chan bool)

	go func() {
		for {
			select {
			case <-shutdown:
				return
			case <-ticker.C:
				cleanup(db)
			}
		}
	}()

	return shutdown
}

func cleanup(db *Database) {
	if err := db.SocialStore.DeleteOldStates(); err != nil {
		slog.Error("failed to delete old states", "err", err)
	}

	if err := db.SessionStore.DeleteOldSessions(); err != nil {
		slog.Error("failed to delete old sessions", "err", err)
	}

	if err := db.OAuthStore.DeleteOldAuthRequests(); err != nil {
		slog.Error("failed to delete old auth requests", "err", err)
	}
}
