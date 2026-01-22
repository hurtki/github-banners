package services

import (
	"context"
	"time"

	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/infrastructure/github"
)

type StatsService struct {
	fetcher *github.Fetcher
	cache   cache.Cache
	ttl     time.Duration
}

func NewStatsService(fetcher *github.Fetcher, cache cache.Cache, ttl time.Duration) *StatsService {
	return &StatsService{
		fetcher: fetcher,
		cache:   cache,
		ttl:     ttl,
	}
}

// GetUserStats returns User Stats with caching
func (s *StatsService) GetUserStats(ctx context.Context, username string) (*domain.GithubUserStats, error) {
	// checking cache for stats
	if cachedStats, found := s.cache.Get(username); found {
		return cachedStats, nil
	}

	userData, err := s.fetcher.FetchUserData(ctx, username)
	if err != nil {
		return nil, err
	}

	//calculating stats from data using the domain logic
	stats := domain.CalculateStats(userData.Repositories)

	// Caching it
	go s.cache.Set(username, &stats, s.ttl)

	return &stats, nil
}

func (s *StatsService) GetMultipleUsers(ctx context.Context, usernames []string) (map[string]*domain.GithubUserStats, error) {
	result := make(map[string]*domain.GithubUserStats)
	errChan := make(chan error, len(usernames))
	statsChan := make(chan struct {
		username string
		stats    *domain.GithubUserStats
	}, len(usernames))

	for _, username := range usernames {
		go func(user string) {
			stats, err := s.GetUserStats(ctx, user)
			if err != nil {
				errChan <- err
				return
			}
			statsChan <- struct {
				username string
				stats    *domain.GithubUserStats
			}{username: user, stats: stats}
		}(username)
	}

	for range usernames {
		select {
		case err := <-errChan:
			_ = err
			// Log error but continue with other users
			continue
		case resultData := <-statsChan:
			result[resultData.username] = resultData.stats
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return result, nil
}
