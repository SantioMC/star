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

	var location string
	geolocation, err := server.GeoService.GetLocation(ctx, address)
	if err != nil {
		slog.Warn("failed to get geolocation data", "err", err, "ip", address)
		location = geolocation.String()
	}

	var accountId string
	if options.Account != nil {
		accountId = options.Account.ID
	}

	session := &database.Session{
		ID:          database.CreateID("sess"),
		UserID:      options.User.ID,
		AccountID:   &accountId,
		Address:     address,
		UserAgent:   userAgent,
		Geolocation: &location,
		Token:       CreateSessionToken(),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(options.Duration),
	}

	return session, nil
}
