package render

import "errors"

var (
	ErrInvalidUsername   = errors.New("invalid username: cannot be empty")
	ErrInvalidUrlPath    = errors.New("invalid url path: cannot be empty")
	ErrInvalidBannerType = errors.New("invalid banner type: template not supported")
	ErrRenderFailure     = errors.New("render failure: unable to generate banner")
	ErrStorageFailure    = errors.New("storage failure: unable to save banner")
)
