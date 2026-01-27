package userstats

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

const (
	SoftTTL = 10 * time.Minute
	HardTTL = 24 * time.Hour
)

type UserStatsService struct {
	repo    GithubStatsRepository
	fetcher UserDataFetcher
	cache   Cache
	cfg     WorkerConfig
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

// WorkerResult holds the result of processing a user
type WorkerResult struct {
	Username string
	Error    error
}
