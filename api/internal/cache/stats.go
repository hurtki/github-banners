package cache

import (
	"strings"
	"time"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/patrickmn/go-cache"
)

// StatsMemoryCache is in memory cache for storing statistics for username
// It uses lower case of username as cache key, so hurtki and HURTKI are same
type StatsMemoryCache struct {
	cache *cache.Cache
}

func NewStatsMemoryCache(defaultTTL time.Duration) *StatsMemoryCache {
	return &StatsMemoryCache{
		cache: cache.New(defaultTTL, 10*time.Minute),
	}
}

func (c *StatsMemoryCache) Get(username string) (*userstats.CachedStats, bool) {
	normalizedUsername := strings.ToLower(username)

	if item, found := c.cache.Get(normalizedUsername); found {
		if stats, ok := item.(*userstats.CachedStats); ok {
			return stats, true
		}
	}
	return nil, false
}

func (c *StatsMemoryCache) Set(username string, entry *userstats.CachedStats, ttl time.Duration) {
	normalizedUsername := strings.ToLower(username)
	c.cache.Set(normalizedUsername, entry, ttl)
}

func (c *StatsMemoryCache) Delete(username string) {
	c.cache.Delete(username)
}
