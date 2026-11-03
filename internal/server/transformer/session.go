package transformer

import (
	"time"

	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/useragent"
)

type UserSession struct {
	ID          string                `json:"id"`
	Address     string                `json:"ip_address" db:"ip_address"`
	UserAgent   string                `json:"user_agent" db:"user_agent"`
	Geolocation *string               `json:"geolocation"`
	Status      database.SessionState `json:"status"`
	Device      string                `json:"device"`
	Browser     string                `json:"browser"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time             `json:"expires_at" db:"expires_at"`
}

func TransformSession(session database.Session) UserSession {
	return UserSession{
		ID:          session.FamilyID,
		Address:     session.Address,
		UserAgent:   session.UserAgent,
		Geolocation: session.Geolocation,
		Status:      session.Status,
		Device:      session.Device,
		Browser:     string(useragent.GetBrowser(session.UserAgent)),
		CreatedAt:   session.CreatedAt,
		ExpiresAt:   session.ExpiresAt,
	}
}

func TransformSessions(sessions []database.Session) []UserSession {
	serialized := []UserSession{}

	for _, value := range sessions {
		session := TransformSession(value)
		if session.Status == database.SessionStateRotated {
			continue
		}

		serialized = append(serialized, session)
	}

	return serialized
}
