package preview

import "errors"

var (
	ErrInvalidBannerType = errors.New("invalid banner type")
	ErrUserDoesntExist   = errors.New("github user doesn't exist")
	ErrInvalidInputs     = errors.New("invalid inputs")
	ErrCantGetPreview    = errors.New("can't get preview")
)
