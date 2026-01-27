package userstats

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

const (
	SoftTTL = 10 * time.Minute
	HardTTL = 24 * time.Hour
)

func NewUserStatsService(repo GithubUserDataRepository, fetcher UserDataFetcher, cache Cache) *UserStatsService {
	return &UserStatsService{
		repo:    repo,
		fetcher: fetcher,
		cache:   cache,
	}
}

func (s *UserStatsService) GetStats(ctx context.Context, username string) (domain.GithubUserStats, error) {
	cached, found := s.cache.Get(username)
	if found {
		//fresh <10mins
		age := time.Since(cached.UpdatedAt)
		if age <= SoftTTL {
			return cached.Stats, nil
		}
		//stalte >10mins but <24 hours
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, _ = s.RecalculateAndSync(bgCtx, username)
		}()
		return cached.Stats, nil
	}

	//checking database if cache missed
	dbData, err := s.repo.GetUserData(context.TODO(), username)
	if err == nil {
		stats := CalculateStats(dbData.Repositories)
		s.cache.Set(username, &CachedStats{
			Stats:     stats,
			UpdatedAt: time.Now(),
		}, HardTTL)
		return stats, nil
	}

	return s.RecalculateAndSync(ctx, username)
}

// fetch api -> save db -> calc stats -> write cache
func (s *UserStatsService) RecalculateAndSync(ctx context.Context, username string) (domain.GithubUserStats, error) {
	//fetching raw data from github
	data, err := s.fetcher.FetchUserData(ctx, username)
	if err != nil {
		return domain.GithubUserStats{}, err
	}

	stats := CalculateStats(data.Repositories)
	//updating database with the new raw data
	if err := s.repo.SaveUserData(context.TODO(), *data); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return domain.GithubUserStats{}, err
		}
		// TODO, we really need to somehow log that we can't save to database or so something with it
	}
	s.cache.Set(username, &CachedStats{
		Stats:     stats,
		UpdatedAt: time.Now(),
	}, HardTTL)

	return stats, nil
}

func (s *UserStatsService) RefreshAll(ctx context.Context, cfg WorkerConfig) (<-chan string, <-chan error) {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}

	jobs := make(chan string, cfg.Concurrency)
	results := make(chan string, cfg.BatchSize)
	errs := make(chan error, cfg.Concurrency*2)

	var workers sync.WaitGroup
	var producer sync.WaitGroup

	workers.Add(cfg.Concurrency)
	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case username, ok := <-jobs:
					if !ok {
						return
					}
					_, err := s.RecalculateAndSync(ctx, username)
					if err != nil {
						errs <- fmt.Errorf("worker: failed to update for %s: %w", username, err)
						continue
					}

					results <- username
				}
			}
		}()
	}

	producer.Go(func() {
		defer close(jobs)

		usernames, err := s.repo.GetAllUsernames(ctx)
		if err != nil {
			errs <- fmt.Errorf("producer: %w", err)
			return
		}
		for _, u := range usernames {
			select {
			case <-ctx.Done():
				return
			case jobs <- u:
			}
		}
	})

	//cleanup
	go func() {
		producer.Wait()
		workers.Wait()
		close(results)
		close(errs)
	}()

	return results, errs
}
