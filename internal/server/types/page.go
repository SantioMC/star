package types

import "github.com/tubblelive/star/internal/database"

type Page[T any] struct {
	Items []T     `json:"items"`
	Next  *string `json:"next"`
}

func NewPage[T any](rows []T, size int, cursor func(T) database.Cursor) Page[T] {
	if len(rows) <= size {
		return Page[T]{Items: rows}
	}

	items := rows[:size]
	next := cursor(items[size-1]).String()

	return Page[T]{Items: items, Next: &next}
}
