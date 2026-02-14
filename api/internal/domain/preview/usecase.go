package preview

import (
	"context"
	"errors"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type StatsService interface {
	GetStats(context.Context, string) (domain.GithubUserStats, error)
}

type PreviewProvider interface {
	GetPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error)
}

type PreviewUsecase struct {
	stats           StatsService
	previewProvider PreviewProvider
}

func NewPreviewUsecase(stats StatsService, previewProvider PreviewProvider) *PreviewUsecase {
	return &PreviewUsecase{
		stats:           stats,
		previewProvider: previewProvider,
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
		switch {
		// if state is not found, then there is no user with this username, returning error
		case errors.Is(err, domain.ErrNotFound):
			return nil, ErrUserDoesntExist
		// if stats service are now unavailable, returning error, that we can't get preview
		case errors.Is(err, domain.ErrUnavailable):
			return nil, ErrCantGetPreview
		default:
			return nil, ErrCantGetPreview
		}
	}

	preview, err := u.previewProvider.GetPreview(ctx, domain.BannerInfo{
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
