package banner

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/hurtki/github-banners/storage/internal/domain"
)

type BannerStorage interface {
	Save(ctx context.Context, name string, extension domain.BannerExtension, content []byte) error
}

type BannerUsecase struct {
	storage BannerStorage
}

func NewBannerUsecase(storage BannerStorage) *BannerUsecase {
	return &BannerUsecase{
		storage: storage,
	}
}

func (u *BannerUsecase) Save(ctx context.Context, in SaveIn) (SaveOut, error) {
	if in.UrlPath == "" || (url.PathEscape(in.UrlPath) != in.UrlPath) {
		return SaveOut{}, ErrInvalidUrlPath
	}
	ext, ok := domain.BannerExtensions[in.Format]

	if !ok {
		return SaveOut{}, ErrInvalidBannerFormat
	}
	err := u.storage.Save(ctx, in.UrlPath, ext, in.BannerData)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnavailable):
			return SaveOut{}, ErrCantSaveBanner
		default:
			return SaveOut{}, fmt.Errorf("%s:%w:%w", "unhandled error from storage", err, ErrCantSaveBanner)
		}
	}

	// returning relative path
	return SaveOut{BannerUrl: path.Join("/banners/", in.UrlPath)}, nil
}
