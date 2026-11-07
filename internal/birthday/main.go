package birthday

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tubblelive/star/internal/database/models"
)

var (
	ErrInvalid       = fiber.NewError(400, "Your birthday must be a valid date in the past")
	ErrRequired      = fiber.NewError(400, "Your birthday is required to use this platform")
	ErrAgeRestricted = fiber.NewError(403, "You do not meet the age requirement to use this platform")
)

type Policy struct {
	Tracked    bool
	MinimumAge int
}

func (p Policy) Enabled() bool {
	return p.Tracked && p.MinimumAge > 0
}

func (p Policy) Check(date *models.Date) error {
	if !p.Enabled() {
		return nil
	}

	if date == nil {
		return ErrRequired
	}

	if date.AddDate(p.MinimumAge, 0, 0).After(time.Now().UTC()) {
		return ErrAgeRestricted
	}

	return nil
}

func Parse(value *string) (*models.Date, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	parsed, err := models.ParseDate(*value)
	if err != nil {
		return nil, ErrInvalid
	}

	now := time.Now().UTC()
	if parsed.After(now) || parsed.Before(now.AddDate(-150, 0, 0)) {
		return nil, ErrInvalid
	}

	return &parsed, nil
}
