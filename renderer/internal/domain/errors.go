package domain

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrUnavailable = errors.New("service unavailable")
)

type ConflictError struct {
	Field ConflictField
}

type ConflictField string

const (
	UnknownConflictField = "unknown conflict field"
)

func NewConflictError(cf ConflictField) error {
	return &ConflictError{
		Field: cf,
	}
}

func (e *ConflictError) Error() string {
	return "conflict"
}
