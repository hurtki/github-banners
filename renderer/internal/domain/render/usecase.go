package render

import (
	"context"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/layout"
)

type BannerStorage interface {
	SaveBanner(ctx context.Context, bannerID string, svg string) (string, error)
}

type BannerRenderer interface {
	RenderBanner(view *layout.BannerView) ([]byte, error)
}

type Usecase struct {
	renderer BannerRenderer
	storage  BannerStorage
}

func NewUsecase(r BannerRenderer, s BannerStorage) *Usecase {
	return &Usecase{
		renderer: r,
		storage:  s,
	}
}

func (u *Usecase) ProcessBanner(ctx context.Context, req UpdateBannerIn) error {
	ltInfo, err := u.validateUpdateBannerIn(req)
	if err != nil {
		return err
	}

	view := layout.BuildView(ltInfo.BannerInfo)

	renderedData, err := u.renderer.RenderBanner(view)
	if err != nil {
		return err
	}

	_, err = u.storage.SaveBanner(ctx, ltInfo.URLPath, string(renderedData))
	if err != nil {
		return err
	}
	return nil
}

func (u *Usecase) validateUpdateBannerIn(req UpdateBannerIn) (domain.LTBannerInfo, error) {
	if req.Username == "" {
		return domain.LTBannerInfo{}, ErrInvalidUsername
	}

	if req.URLPath == "" {
		return domain.LTBannerInfo{}, ErrInvalidUrlPath
	}

	bannerType := domain.BannerTypeDefault

	switch req.BannerType {
	case string(domain.BannerTypeDark):
		bannerType = domain.BannerTypeDark
	case string(domain.BannerTypeDefault):
		bannerType = domain.BannerTypeDefault
	default:
		return domain.LTBannerInfo{}, ErrInvalidBannerType
	}

	return domain.LTBannerInfo{
		URLPath: req.URLPath,
		BannerInfo: domain.BannerInfo{
			Username:   req.Username,
			BannerType: bannerType,
			Stats:      req.Stats,
		},
	}, nil
}

func (u *Usecase) Render(ctx context.Context, req RenderIn) ([]byte, error) {
	info, err := u.validateRenderIn(req)
	if err != nil {
		return nil, err
	}

	view := layout.BuildView(info)
	renderedData, err := u.renderer.RenderBanner(view)
	if err != nil {
		return nil, err
	}
	return renderedData, nil
}

func (u *Usecase) validateRenderIn(req RenderIn) (domain.BannerInfo, error) {
	if req.Username == "" {
		return domain.BannerInfo{}, ErrInvalidUsername
	}

	bannerType := domain.BannerTypeDefault

	switch req.BannerType {
	case string(domain.BannerTypeDark):
		bannerType = domain.BannerTypeDark
	case string(domain.BannerTypeDefault):
		bannerType = domain.BannerTypeDefault
	default:
		return domain.BannerInfo{}, ErrInvalidBannerType
	}

	return domain.BannerInfo{
		Username:   req.Username,
		BannerType: bannerType,
		Stats:      req.Stats,
	}, nil
}
