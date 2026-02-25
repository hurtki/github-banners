package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrUnavailable       = errors.New("service unavailable")

	//Banner specific errors
	ErrInvalidUsername   = errors.New("invalid username: cannot be empty")
	ErrInvalidBannerType = errors.New("invalid banner type: template not supported")
	ErrRenderFailure     = errors.New("render failure: unable to generate banner")
	ErrStorageFailure    = errors.New("storage failure: unable to save banner")
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
