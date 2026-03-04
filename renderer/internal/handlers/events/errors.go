package events

import "errors"

var (
	ErrValidation = errors.New("validation error")
	ErrBusiness   = errors.New("business error")
	ErrTransient  = errors.New("transient error")
	ErrDuplicate  = errors.New("duplicate event")
)
