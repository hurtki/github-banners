package usecase

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/hurtki/github-banners/storage/internal/domain"
)

const (
	storageMSBannersStorageUrl = "http://localhost/banners/"
)

type BannerStorage interface {
	Save(name string, extension domain.BannerExtension, content []byte) error
}

type BannerUsecase struct {
	storage BannerStorage
}

func NewBannerUsecase(storage BannerStorage) *BannerUsecase {
	return &BannerUsecase{
		storage: storage,
	}
}

func (u *BannerUsecase) Save(in SaveIn) (SaveOut, error) {
	if in.UrlPath == "" || (url.PathEscape(in.UrlPath) != in.UrlPath) {
		return SaveOut{}, ErrInvalidUrlPath
	}
	ext, ok := domain.BannerExtensions[in.Format]

	if !ok {
		return SaveOut{}, ErrInvalidBannerFormat
	}
	err := u.storage.Save(in.UrlPath, ext, in.BannerData)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnavailable):
			return SaveOut{}, ErrCantSaveBanner
		default:
			return SaveOut{}, fmt.Errorf("%s:%w", "unhandled error from storage", err)
		}
	}
	return SaveOut{BannerUrl: storageMSBannersStorageUrl + in.UrlPath}, nil
}
