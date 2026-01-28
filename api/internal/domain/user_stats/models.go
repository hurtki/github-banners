package userstats

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type UserStatsService struct {
	repo    GithubUserDataRepository
	fetcher UserDataFetcher
	cache   Cache
}

type CachedStats struct {
	Stats     domain.GithubUserStats
	UpdatedAt time.Time
}

type WorkerConfig struct {
	BatchSize   int
	Concurrency int
	CacheTTL    time.Duration
}
