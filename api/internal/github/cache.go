package github

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache interface {
	Get(username string) (*UserStats, bool)
	Set(username string, stats *UserStats)
	Delete(username string)
}

type MemoryCache struct {
	cache *cache.Cache
}

func NewCache(defaultTTL time.Duration) Cache {
	return &MemoryCache{
		cache: cache.New(defaultTTL, 10*time.Minute),
	}
}

func (c *MemoryCache) Get(username string) (*UserStats, bool) {
	if item, found := c.cache.Get(username); found {
		return item.(*UserStats), true
	}
	return nil, false
}

func (c *MemoryCache) Set(username string, stats *UserStats) {
	c.cache.Set(username, stats, cache.DefaultExpiration)
}

func (c *MemoryCache) Delete(username string) {
	c.cache.Delete(username)
}
