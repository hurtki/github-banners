package longterm

import "errors"

var (
	ErrInvalidBannerType   = errors.New("invalid banner type")
	ErrUserDoesntExist     = errors.New("github user doesn't exist")
	ErrInvalidInputs       = errors.New("invalid inputs")
	ErrBannerAlreadyExists = errors.New("banner already exists")
	ErrCantCreateBanner    = errors.New("can't create banner")
)
