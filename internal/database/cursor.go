package database

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Cursor struct {
	CreatedAt time.Time
	ID        string
}

func (cursor Cursor) String() string {
	return strconv.FormatInt(cursor.CreatedAt.UnixMicro(), 10) + "." + cursor.ID
}

func ParseCursor(value string) (*Cursor, error) {
	micros, id, found := strings.Cut(value, ".")
	if !found || id == "" {
		return nil, errors.New("malformed cursor")
	}

	parsed, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return nil, err
	}

	return &Cursor{CreatedAt: time.UnixMicro(parsed), ID: id}, nil
}
