package types

import "time"

type MetricsResponse struct {
	Users          int       `json:"users"`
	ActiveSessions int       `json:"active_sessions"`
	Applications   int       `json:"applications"`
	OpenAlerts     int       `json:"open_alerts"`
	Version        string    `json:"version"`
	StartedAt      time.Time `json:"started_at"`
}
