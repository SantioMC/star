package session

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
)

var ErrUserNotFound = errors.New("user must be specified in session options")

func CreateSession(
	ctx fiber.Ctx,
	server *state.StarServer,
	options *Options,
) (*database.Session, error) {
	if options.User == nil {
		return nil, ErrUserNotFound
	}

	address := ctx.IP()
	userAgent := ctx.Get(fiber.HeaderUserAgent)

	location, err := server.GeoService.GetLocation(ctx, address)
	if err != nil {
		slog.Warn("failed to get geolocation data", "err", err, "ip", address)
	}

	var accountId string
	if options.Account != nil {
		accountId = options.Account.ID
	}

	device := parseUserAgent(userAgent)

	session := &database.Session{
		ID:            database.CreateID(database.SessionKey),
		UserID:        options.User.ID,
		AccountID:     &accountId,
		Address:       address,
		UserAgent:     userAgent,
		Geolocation:   location,
		Device:        string(device),
		Token:         CreateSessionToken(),
		ApplicationID: server.DefaultApp.ID,
		Status:        database.SessionStateActive,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(options.Duration),
	}

	if err := server.Database.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}
