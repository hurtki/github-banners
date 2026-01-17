package repo

import (
	"errors"
	"fmt"
)

// ErrConflictValue represents error, when some of the values your are trying to add conflicts with storage ruleset
type ErrConflictValue struct{ Field string }

func (e ErrConflictValue) Error() string {
	return "conflict, there is already a value, that should be unique in " + e.Field
}

// ErrEmptyField represents error, when some of values you are adding is Empty and that conflicts with storage ruleset
type ErrEmptyField struct{ Field string }

func (e ErrEmptyField) Error() string { return fmt.Sprintf("field %s should be not empty", e.Field) }

// ErrRepoInternal represents error, when storage faced internal problem, not connected to the inputs
type ErrRepoInternal struct{ Note string }

func (e ErrRepoInternal) Error() string {
	return fmt.Sprintf("internal repo occured, note: %s", e.Note)
}

var (
	// ErrNothingChanged apears, when since executing "repo request", nothing changed
	ErrNothingChanged = errors.New("nothing changed")
	// ErrNothingFound apears, when nothing was found for your "repo request"
	ErrNothingFound = errors.New("nothing found")
)
