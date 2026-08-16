package token

import "time"

type Options struct {
	Duration    time.Duration
	Permissions []string
	Scope       []string
	SuperUser   bool
}
