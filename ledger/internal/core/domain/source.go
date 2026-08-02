package domain

import "errors"

// ErrDuplicate is returned when a unique constraint is violated.
var ErrDuplicate = errors.New("duplicate entry")

const (
	SourceKindBank = "bank"
	SourceKindCard = "card"
)

type Source struct {
	ID   int
	Name string
	Kind string
}
