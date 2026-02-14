package cache

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/patrickmn/go-cache"
)

// PreviewMemoryCache is used to cache rendered banners with same domain.BannerInfo
// Get returns banner ( nil if found is false); bannerInfo hash ( always returned, same for same BannerInfo ); found ( is banner found in cache )
// Set uses bannerInfo hash, that Get returned to store rendered banner in cache
type PreviewMemoryCache struct {
	cache       *cache.Cache
	ttl         time.Duration
	hashCounter bannerInfoHashCounter
}

func NewPreviewMemoryCache(ttl time.Duration) *PreviewMemoryCache {
	return &PreviewMemoryCache{
		cache:       cache.New(ttl, time.Minute*10),
		ttl:         ttl,
		hashCounter: newBannerInfoHashCounter(),
	}
}

// Get gets rendered banner from cache and returns it
// Second return is hash, that will be the same for same bannerInfo ( excluding FetchedAt field, for different FetchedAt fields it will be same, if other fields are the same)
// Third return is found, if found is false, then banner pointer is nil ( but hash is valid )
func (c *PreviewMemoryCache) Get(bf domain.BannerInfo) (*domain.Banner, string, bool) {
	hashKey := c.hashCounter.Hash(bf)
	if item, found := c.cache.Get(hashKey); found {
		if banner, ok := item.(*domain.Banner); ok {
			return banner, hashKey, true
		}
	}
	return nil, hashKey, false
}

// Set sets rendered banner to cache using hash, that Get method generated
// banner pointer shouldn't be nil, otherwise panic
func (c *PreviewMemoryCache) Set(hashKey string, banner *domain.Banner) {
	if banner == nil {
		panic("nil banner set in PreviewMemoryCache")
	}
	c.cache.Set(hashKey, banner, c.ttl)
}
