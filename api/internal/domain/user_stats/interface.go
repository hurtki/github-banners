package userstats

import (
	"context"
	"github.com/hurtki/github-banners/api/internal/domain"
	"time"
)

type GithubStatsRepository interface {
	GetUserData(username string) (domain.GithubUserData, error)
	UpdateUserData(userData domain.GithubUserData) error
	GetAllUsernames() ([]string, error)
}

type Cache interface {
	Get(username string) (*CachedStats, bool)
	Set(username string, entry *CachedStats, ttl time.Duration)
	Delete(username string)
}

type UserDataFetcher interface {
	FetchUserData(ctx context.Context, username string) (*domain.GithubUserData, error)
}

type CacheWriter interface {
	Set(username string, stats *domain.GithubUserStats, ttl time.Duration)
}

// NewUserStatsService wires the domain service with its required ports.
func NewUserStatsService(repo GithubStatsRepository, fetcher UserDataFetcher, cache Cache, cfg WorkerConfig) *UserStatsService {
	return &UserStatsService{
		repo:    repo,
		fetcher: fetcher,
		cache:   cache,
		cfg:     cfg,
	}
}
