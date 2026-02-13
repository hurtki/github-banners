package preview

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/domain"
	"golang.org/x/sync/singleflight"
)

type PreviewRenderer interface {
	RenderPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error)
}

type Cache interface {
	Get(domain.BannerInfo) (*domain.Banner, string, bool)
	Set(string, *domain.Banner)
}

// PreviewService is a service that caches renderer results
// Consider it as a caching wrap for renderer infrastrcture
// At GetPreview method it will return same result for same bannerInfo, if cache is still valid
type PreviewService struct {
	renderer PreviewRenderer
	cache    Cache
	g        singleflight.Group
}

func NewPreviewService(renderer PreviewRenderer, cache Cache) *PreviewService {
	return &PreviewService{
		renderer: renderer,
		cache:    cache,
		g:        singleflight.Group{},
	}
}

func (s *PreviewService) GetPreview(ctx context.Context, bannerInfo domain.BannerInfo) (*domain.Banner, error) {
	banner, hash, ok := s.cache.Get(bannerInfo)
	if ok {
		return banner, nil
	}

	res, err, ok := s.g.Do(hash, func() (any, error) {
		return s.renderer.RenderPreview(ctx, bannerInfo)
	})

	if err != nil {
		return nil, err
	}

	if banner, ok := res.(*domain.Banner); ok {
		s.cache.Set(hash, banner)
		return banner, nil
	}
	return nil, domain.ErrUnavailable
}
