package cache

import (
	"time"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/patrickmn/go-cache"
)

type StatsMemoryCache struct {
	cache *cache.Cache
}

func NewStatsMemoryCache(defaultTTL time.Duration) *StatsMemoryCache {
	return &StatsMemoryCache{
		cache: cache.New(defaultTTL, 10*time.Minute),
	}
}

func (c *StatsMemoryCache) Get(username string) (*userstats.CachedStats, bool) {
	if item, found := c.cache.Get(username); found {
		if stats, ok := item.(*userstats.CachedStats); ok {
			return stats, true
		}
	}
	return nil, false
}

func (c *StatsMemoryCache) Set(username string, entry *userstats.CachedStats, ttl time.Duration) {
	c.cache.Set(username, entry, ttl)
}

func (c *StatsMemoryCache) Delete(username string) {
	c.cache.Delete(username)
}
