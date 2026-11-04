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
var ErrUserSuspended = fiber.NewError(403, "This account has been suspended by the platform")

func rejectSuspended(ctx fiber.Ctx, server *state.StarServer, userID string) error {
	suspension, err := server.Database.GetActiveSuspension(ctx, userID)
	if err != nil {
		return err
	} else if suspension != nil {
		return ErrUserSuspended
	}

	return nil
}

func CreateSession(
	ctx fiber.Ctx,
	server *state.StarServer,
	options *Options,
) (*database.Session, error) {
	if options.User == nil {
		return nil, ErrUserNotFound
	}

	if err := rejectSuspended(ctx, server, options.User.ID); err != nil {
		return nil, err
	}

	var accountID *string
	if options.Account != nil {
		accountID = &options.Account.ID
	}

	applicationID := server.DefaultApp.Load().ID
	if options.Application != nil {
		applicationID = options.Application.ID
	}

	authenticatedAt := options.AuthenticatedAt
	if authenticatedAt.IsZero() {
		authenticatedAt = time.Now()
	}

	session := buildSession(
		ctx,
		server,
		database.CreateID(database.SessionFamilyKey),
		options.User.ID,
		accountID,
		applicationID,
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
	familyID string,
	userId string,
	accountId *string,
	applicationId string,
	scopes pq.StringArray,
	duration time.Duration,
	authenticatedAt time.Time,
) *database.Session {
	address := server.ClientIP(ctx)
	userAgent := ctx.Get(fiber.HeaderUserAgent)

	location, err := server.GeoService.GetLocation(ctx, address)
	if err != nil {
		slog.Warn("failed to get geolocation data", "err", err, "ip", address)
	}

	return &database.Session{
		ID:              database.CreateID(database.SessionKey),
		FamilyID:        familyID,
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
