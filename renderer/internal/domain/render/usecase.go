package render

import (
	"context"

	"github.com/hurtki/github-banners/renderer/internal/domain"
)

type BannerStorage interface {
	SaveBanner(ctx context.Context, bannerID string, svg string) (string, error)
}

type Usecase struct {
	renderer *Service
	storage  BannerStorage
}

func NewUsecase(r *Service, s BannerStorage) *Usecase {
	return &Usecase{
		renderer: r,
		storage:  s,
	}
}

func (u *Usecase) Execute(ctx context.Context, req domain.BannerInfo) error {
	if err := u.validate(req); err != nil {
		return err
	}
	renderedData, err := u.renderer.Render(req)
	if err != nil {
		return err
	}

	_, err = u.storage.SaveBanner(ctx, req.URLPath, string(renderedData))
	if err != nil {
		return err
	}
	return nil
}

func (u *Usecase) validate(info domain.BannerInfo) error {
	if info.Username == "" {
		return ErrInvalidUsername
	}

	if info.URLPath == "" {
		return ErrInvalidUrlPath
	}

	switch info.BannerType {
	case domain.BannerTypeDefault, domain.BannerTypeDark:
		// valid
	default:
		return ErrInvalidBannerType
	}
	return nil
}
