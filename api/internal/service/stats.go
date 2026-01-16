package service

import (
	"context"

	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/infrastructure/github"
)

type StatsService struct {
	fetcher *github.Fetcher
	cache   cache.Cache
}

func NewStatsService(fetcher *github.Fetcher, cache cache.Cache) *StatsService {
	return &StatsService{
		fetcher: fetcher,
		cache:   cache,
	}
}

// GetUserStats returns user stats with caching
func (s *StatsService) GetUserStats(ctx context.Context, username string) (*domain.UserStats, error) {
	// checking cache first
	if cachedStats, found := s.cache.Get(username); found {
		cachedStats.Cached = true
		return cachedStats, nil
	}

	stats, err := s.fetcher.FetchUserStats(ctx, username)
	if err != nil {
		return nil, err
	}

	//calculating status using the domain logic
	stats.Stats = domain.CalculateStats(stats.Repositories)

	// Cache it
	go s.cache.Set(username, stats)

	return stats, nil
}

func (s *StatsService) GetMultipleUsers(ctx context.Context, usernames []string) (map[string]*domain.UserStats, error) {
	result := make(map[string]*domain.UserStats)
	errChan := make(chan error, len(usernames))
	statsChan := make(chan struct {
		username string
		stats    *domain.UserStats
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
				stats    *domain.UserStats
			}{username: user, stats: stats}
		}(username)
	}

	for i := 0; i < len(usernames); i++ {
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
