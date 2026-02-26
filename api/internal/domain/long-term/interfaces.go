package longterm

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type BannerRepo interface {
	GetActiveBanners(ctx context.Context) ([]domain.LTBannerMetadata, error)
	SaveBanner(ctx context.Context, banner domain.LTBannerMetadata) error
	DeactivateBanner(ctx context.Context, githubUsername string, bannerType domain.BannerType) error
	GetBanner(ctx context.Context, githubUsername string, bannerType domain.BannerType) (domain.LTBannerMetadata, error)
}

type StatsService interface {
	GetStats(context.Context, string) (domain.GithubUserStats, error)
}

type UpdateRequestPublisher interface {
	Publish(ctx context.Context, info domain.LTBannerInfo) error
}

type PreviewService interface {
	GetPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error)
}

type StorageClient interface {
	SaveBanner(ctx context.Context, urlPath string, svg string) (string, error)
}
