package longterm

import "errors"

var (
	ErrInvalidBannerType = errors.New("invalid banner type")
	ErrUserDoesntExist   = errors.New("github user doesn't exist")
	ErrCantCreateBanner  = errors.New("can't create banner")
)
