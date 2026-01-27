package userstats

import (
	"context"
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type GithubUserDataRepository interface {
	SaveUserData(ctx context.Context, userData domain.GithubUserData) error
	GetUserData(ctx context.Context, username string) (domain.GithubUserData, error)
	GetAllUsernames(ctx context.Context) ([]string, error)
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
