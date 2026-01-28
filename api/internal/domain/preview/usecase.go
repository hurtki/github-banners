package preview

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
)

type StatsService interface {
	GetStats(context.Context, string) (domain.GithubUserStats, error)
}

type BannerPreview struct {
	SvgData []byte
}

type PreviewRenderer interface {
	RenderPreview(ctx context.Context, bannerInfo renderer.GithubUserBannerInfo) (BannerPreview, error)
}
