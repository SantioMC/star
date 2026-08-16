package session

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/state"
	"github.com/tubblelive/star/internal/useragent"
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

	var accountId *string
	if options.Account != nil {
		accountId = &options.Account.ID
	}

	applicationId := server.DefaultApp.ID
	if options.Application != nil {
		applicationId = options.Application.ID
	}

	authenticatedAt := options.AuthenticatedAt
	if authenticatedAt.IsZero() {
		authenticatedAt = time.Now()
	}

	session := buildSession(
		ctx,
		server,
		options.User.ID,
		accountId,
		applicationId,
		pq.StringArray(options.Scopes),
		options.Duration,
		authenticatedAt,
	)

	if err := server.Database.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func buildSession(
	ctx fiber.Ctx,
	server *state.StarServer,
	userId string,
	accountId *string,
	applicationId string,
	scopes pq.StringArray,
	duration time.Duration,
	authenticatedAt time.Time,
) *database.Session {
	address := ctx.IP()
	userAgent := ctx.Get(fiber.HeaderUserAgent)

	location, err := server.GeoService.GetLocation(ctx, address)
	if err != nil {
		slog.Warn("failed to get geolocation data", "err", err, "ip", address)
	}

	return &database.Session{
		ID:              database.CreateID(database.SessionKey),
		UserID:          userId,
		AccountID:       accountId,
		Address:         address,
		UserAgent:       userAgent,
		Geolocation:     location,
		Device:          string(useragent.GetDevice(userAgent)),
		Token:           CreateSessionToken(),
		Scopes:          scopes,
		ApplicationID:   applicationId,
		Status:          database.SessionStateActive,
		AuthenticatedAt: authenticatedAt,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(duration),
	}
}
