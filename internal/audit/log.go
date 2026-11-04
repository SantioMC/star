package audit

import (
	"encoding/json"
	"errors"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database"
	"github.com/tubblelive/star/internal/server/common"
	"github.com/tubblelive/star/internal/server/request"
	"github.com/tubblelive/star/internal/state"
)

type record struct {
	database.AuditLog
	anonymize bool
}

type Option func(*record)

func Log(ctx fiber.Ctx, action Action, options ...Option) {
	star := common.GetState(ctx)

	entry := &record{
		AuditLog: database.AuditLog{
			ID:        database.CreateID(database.AuditLogKey),
			Action:    string(action),
			CreatedAt: time.Now(),
		},
		anonymize: star.Settings.Load().AuditAnonymize,
	}

	for _, option := range options {
		option(entry)
	}

	entry.resolveActor(ctx, star)

	if !entry.anonymize {
		entry.Address = new(star.ClientIP(ctx))
		entry.UserAgent = userAgent(ctx)
	}

	if err := write(ctx, star, &entry.AuditLog); err != nil {
		slog.Error("failed to write audit log", "action", action, "err", err)
	}
}

func write(ctx fiber.Ctx, star *state.StarServer, log *database.AuditLog) error {
	var err error

	for attempt := range 3 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}

		err = star.Database.CreateAuditLog(ctx, log)
		if err == nil || errors.Is(err, database.ErrDuplicate) {
			return nil
		}
	}

	return err
}

func (entry *record) resolveActor(ctx fiber.Ctx, star *state.StarServer) {
	if jwt, err := request.GetJWT(ctx); err == nil {
		if entry.ActorID == nil {
			entry.ActorID = &jwt.Subject
		}

		if entry.Application == nil {
			entry.Application = applicationID(ctx, star, jwt.ClientID)
		}
	}

	if entry.anonymize || entry.ActorID == nil || entry.ActorName != nil {
		return
	}

	if user := star.Database.GetUserByID(ctx, *entry.ActorID); user != nil {
		entry.ActorName = &user.Name
	}
}

func applicationID(ctx fiber.Ctx, star *state.StarServer, clientID string) *string {
	if first := star.DefaultApp.Load(); first.ClientID == clientID {
		return &first.ID
	}

	if application := star.Database.GetApplicationByClientID(ctx, clientID); application != nil {
		return &application.ID
	}

	return nil
}

func userAgent(ctx fiber.Ctx) *string {
	agent := ctx.Get(fiber.HeaderUserAgent)
	if agent == "" {
		return nil
	}

	if len(agent) > 512 {
		agent = agent[:512]
	}

	return &agent
}

func (entry *record) set(key string, value any) {
	if entry.Metadata == nil {
		entry.Metadata = map[string]any{}
	}

	entry.Metadata[key] = value
}

func By(user *database.User) Option {
	return func(entry *record) {
		entry.ActorID = &user.ID

		if !entry.anonymize {
			entry.ActorName = &user.Name
		}
	}
}

func ByID(userID string) Option {
	return func(entry *record) {
		entry.ActorID = &userID
	}
}

func On(kind TargetType, id string, name string) Option {
	return func(entry *record) {
		entry.TargetType = new(string(kind))
		entry.TargetID = &id

		if name != "" {
			entry.TargetName = &name
		}

		switch kind {
		case TargetApplication:
			entry.Application = &id
		case TargetProvider:
			entry.Provider = &id
		}
	}
}

func ViaProvider(providerID string) Option {
	return func(entry *record) {
		entry.Provider = &providerID
	}
}

func OnUser(user *database.User) Option {
	return func(entry *record) {
		entry.TargetType = new(string(TargetUser))
		entry.TargetID = &user.ID

		if !entry.anonymize {
			entry.TargetName = &user.Name
		}
	}
}

func With(key string, value any) Option {
	return func(entry *record) {
		entry.set(key, value)
	}
}

func Private(key string, value any) Option {
	return func(entry *record) {
		if !entry.anonymize {
			entry.set(key, value)
		}
	}
}

func Changes(before any, after any) Option {
	return func(entry *record) {
		if fields := changedFields(before, after); len(fields) > 0 {
			entry.set("fields", fields)
		}
	}
}

func changedFields(before any, after any) []string {
	previous, err := fields(before)
	if err != nil {
		slog.Error("failed to compare audit log fields", "err", err)
		return nil
	}

	current, err := fields(after)
	if err != nil {
		slog.Error("failed to compare audit log fields", "err", err)
		return nil
	}

	changed := []string{}
	for _, key := range slices.Sorted(maps.Keys(current)) {
		if string(previous[key]) != string(current[key]) {
			changed = append(changed, key)
		}
	}

	return changed
}

func fields(model any) (map[string]json.RawMessage, error) {
	encoded, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	var decoded map[string]json.RawMessage
	err = json.Unmarshal(encoded, &decoded)

	return decoded, err
}
