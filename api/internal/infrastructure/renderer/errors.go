package renderer

import "errors"

var (
	// internal package errors
	ErrCantRequestRenderer = errors.New("can't request renderer")
	ErrBadPreviewRequest   = errors.New("bad preview request")
)
