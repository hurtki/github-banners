package usecase

import (
	"context"

	"github.com/hurtki/github-banners/renderer/internal/domain"
	"github.com/hurtki/github-banners/renderer/internal/domain/service"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

type RenderedBannerUsecase struct {
	renderer *service.RenderService
	storage  domain.BannerStorage
	log      logger.Logger
}

func NewRenderedBannerUsecase(r *service.RenderService, s domain.BannerStorage, log logger.Logger) *RenderedBannerUsecase {
	return &RenderedBannerUsecase{
		renderer: r,
		storage:  s,
		log:      log.With("service", "render-banner-usecase"),
	}
}

func (u *RenderedBannerUsecase) Execute(ctx context.Context, req domain.BannerInfo) error {
	fn := "usecase.RenderedBannerUsecase.Execute"

	u.log.Debug("processing banner render request", "source", fn, "username", req.Username, "banner_type", req.BannerType)
	if err := u.validate(req); err != nil {
		u.log.With("validation failed", "source", fn, "err", err)
		return err
	}

	renderedBanner, err := u.renderer.Render(req)
	if err != nil {
		u.log.Error("rendering step failed", "source", fn, "err", err)
		return err
	}

	u.log.Debug("attempting to store rendered banner", "source", fn, "filename", renderedBanner.Filename)

	if err := u.storage.Save(ctx, renderedBanner); err != nil {
		u.log.Error("storage operation failed", "source", fn, "err", err, "filename", renderedBanner.Filename)
		return domain.ErrStorageFailure
	}

	u.log.Info("banner successfully rendered and stored", "source", fn, "filename", renderedBanner.Filename)
	return nil
}

func (u *RenderedBannerUsecase) validate(info domain.BannerInfo) error {
	if info.Username == "" {
		return domain.ErrInvalidUsername
	}

	validTypes := map[string]bool{
		"default": true,
		"dark":    true,
	}

	if !validTypes[info.BannerType] {
		return domain.ErrInvalidBannerType
	}
	return nil
}
