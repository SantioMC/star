package token

import (
	"time"

	"github.com/tubblelive/star/internal/database"
)

type Options struct {
	Application *database.Application
	Duration    time.Duration
	Audience    []string
	Permissions []string
	Scope       []string
	Claims      map[string]any
	SuperUser   bool
}
