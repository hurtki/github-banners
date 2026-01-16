package repo

import (
	"errors"
	"fmt"
)

type ErrConflictValue struct{ Field string }

func (e ErrConflictValue) Error() string {
	return "conflict, there is already a value, that should be unique in " + e.Field
}

type ErrEmptyField struct{ Field string }

func (e ErrEmptyField) Error() string { return fmt.Sprintf("field %s should be not empty", e.Field) }

type ErrRepoInternal struct{ Note string }

func (e ErrRepoInternal) Error() string {
	return fmt.Sprintf("internal repo occured, note: %s", e.Note)
}

var (
	ErrNothingChanged = errors.New("nothing changed")
	ErrNothingFound   = errors.New("nothing found")
)
