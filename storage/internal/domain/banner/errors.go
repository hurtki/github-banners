package banner

import (
	"errors"
)

var (
	ErrInvalidUrlPath      = errors.New("invalid url path")
	ErrInvalidBannerFormat = errors.New("invalid banner format")
	ErrCantSaveBanner      = errors.New("cant save banner")
)
