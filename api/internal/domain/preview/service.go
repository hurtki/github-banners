package preview

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type PreviewRenderer interface {
	RenderPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error)
}

type Cache interface {
	Get(domain.BannerInfo) (*domain.Banner, bool)
	Set(domain.BannerInfo, *domain.Banner)
}

// PreviewService is a service that caches renderer results
// Consider it as a caching wrap for renderer infrastrcture
// At GetPreview method it will return same result for same bannerInfo, if cache is still valid
type PreviewService struct {
	renderer PreviewRenderer
	cache    Cache
}

func NewPreviewService(renderer PreviewRenderer, cache Cache) *PreviewService {
	return &PreviewService{
		renderer: renderer,
		cache:    cache,
	}
}

func (s *PreviewService) GetPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error) {
	if banner, ok := s.cache.Get(bannerInfo); ok {
		return banner, nil
	}

	banner, err := s.renderer.RenderPreview(ctx, bannerInfo)
	if err != nil {
		return nil, err
	}
	s.cache.Set(bannerInfo, banner)
	return banner, nil
}
