package cache

import (
	"time"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	userstats.Cache
}

type MemoryCache struct {
	cache *cache.Cache
}

func NewCache(defaultTTL time.Duration) Cache {
	return &MemoryCache{
		cache: cache.New(defaultTTL, 10*time.Minute),
	}
}

func (c *MemoryCache) Get(username string) (*userstats.CachedStats, bool) {
	if item, found := c.cache.Get(username); found {
		if stats, ok := item.(*userstats.CachedStats); ok {
			return stats, true
		}
	}
	return nil, false
}

func (c *MemoryCache) Set(username string, entry *userstats.CachedStats, ttl time.Duration) {
	c.cache.Set(username, entry, ttl)
}

func (c *MemoryCache) Delete(username string) {
	c.cache.Delete(username)
}
