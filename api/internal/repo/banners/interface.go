package banners

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type BannerRepo interface {
	GetActiveBanners(ctx context.Context) ([]domain.LTBannerInfo, error)
	AddBanner(ctx context.Context, banner domain.LTBannerInfo) error
	DeactivateBanner(ctx context.Context, githubUsername string) error
	IsActive(ctx context.Context, githubUsername string) (bool, error)
}
