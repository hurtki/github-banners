package cache

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	Get(username string) (*domain.GithubUserStats, bool)
	Set(username string, stats *domain.GithubUserStats)
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

func (c *MemoryCache) Get(username string) (*domain.GithubUserStats, bool) {
	if item, found := c.cache.Get(username); found {
		return item.(*domain.GithubUserStats), true
	}
	return nil, false
}

func (c *MemoryCache) Set(username string, stats *domain.GithubUserStats) {
	c.cache.Set(username, stats, cache.DefaultExpiration)
}

func (c *MemoryCache) Delete(username string) {
	c.cache.Delete(username)
}
