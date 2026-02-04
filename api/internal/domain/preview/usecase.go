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
	RenderPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error)
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

func (u *PreviewUsecase) GetPreview(ctx context.Context, username string, bannerType string) (*domain.Banner, error) {
	// bannerType validation
	bt, ok := domain.BannerTypes[bannerType]
	if !ok {
		return nil, ErrInvalidBannerType
	}

	// getting user's statisctics
	userStats, err := u.stats.GetStats(ctx, username)

	if err != nil {
		// if state is not found, then there is no user with this username, returning error
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrUserDoesntExist
		}
		// if stats service are now unavailable, returning error, that we can't get preview
		if errors.Is(err, domain.ErrUnavailable) {
			return nil, ErrCantGetPreview
		}
		return nil, ErrCantGetPreview
	}

	preview, err := u.renderer.RenderPreview(ctx, domain.BannerInfo{
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
