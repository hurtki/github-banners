package preview

import (
	"context"
	"errors"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type StatsService interface {
	GetStats(context.Context, string) (domain.GithubUserStats, error)
}

type PreviewRenderer interface {
	RenderPreview(ctx context.Context, bannerInfo domain.GithubUserBannerInfo) (*domain.GithubBanner, error)
}

type PreviewUsecase struct {
	stats    StatsService
	renderer PreviewRenderer
}

func NewPreviewUsecase(stats StatsService, renderer PreviewRenderer) *PreviewUsecase {
	return &PreviewUsecase{
		stats:    stats,
		renderer: renderer,
	}
}

func (u *PreviewUsecase) GetPreview(ctx context.Context, username string, bannerType string) (*domain.GithubBanner, error) {
	// bannerType validation
	bt, ok := domain.BannerTypes[bannerType]
	if !ok {
		return nil, ErrInvalidBannerType
	}

	// getting user's statisctics
	userStats, err := u.stats.GetStats(ctx, username)

	if err != nil {
		// TODO create issue on linear about error handling in github fetches infrastrcture
		return nil, ErrCantGetPreview
	}

	preview, err := u.renderer.RenderPreview(ctx, domain.GithubUserBannerInfo{
		Username:   username,
		BannerType: bt,
		Stats:      userStats,
	})

	if err != nil {
		if err == domain.ErrUnavailable {
			return nil, ErrCantGetPreview
		}
		var cfErr *domain.ConflictError
		if ok := errors.As(err, &cfErr); ok {
			if cfErr.Field == domain.UnknownConflictField {
				return nil, ErrInvalidInputs
			} else {
				return nil, ErrCantGetPreview
			}
		}
		return nil, ErrCantGetPreview
	}

	return preview, nil
}
