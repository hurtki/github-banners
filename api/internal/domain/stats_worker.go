package domain

import (
	"context"
	"fmt"
	"sync"
	"time"
)
// CalculateStats aggregates repository statistics without additional API calls.
func CalculateStats(repos []GithubRepository) GithubUserStats {
	var stats GithubUserStats
	stats.Languages = make(map[string]int)

	for _, repo := range repos {
		if repo.Fork {
			stats.ForkedRepos++
		} else {
			stats.OriginalRepos++
			stats.TotalStars += repo.StarsCount
			stats.TotalForks += repo.ForksCount

			if lang := repo.Language; lang != nil {
				stats.Languages[*lang] += 1
			}
		}
	}
	stats.TotalRepos = len(repos)
	return stats
}

type GithubStatsRepository interface {
	GetUserData(username string) (GithubUserData, error)
	UpdateUserData(userData GithubUserData) error
	GetAllUsernames() ([]string, error)
}

type UserDataFetcher interface {
	FetchUserData(ctx context.Context, username string) (*GithubUserData, error)
}

type CacheWriter interface {
	Set(username string, stats *GithubUserStats, ttl time.Duration)
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

func RunUserStatsWorker(
	ctx context.Context,
	repo GithubStatsRepository,
	fetcher UserDataFetcher,
	cache CacheWriter,
	cfg WorkerConfig,
) (<-chan string, <-chan error) {
	// Validate dependencies
	if repo == nil {
		ch := make(chan error, 1)
		ch <- fmt.Errorf("repository cannot be nil")
		close(ch)
		return nil, ch
	}
	if fetcher == nil {
		ch := make(chan error, 1)
		ch <- fmt.Errorf("fetcher cannot be nil")
		close(ch)
		return nil, ch
	}
	if cache == nil {
		ch := make(chan error, 1)
		ch <- fmt.Errorf("cache cannot be nil")
		close(ch)
		return nil, ch
	}

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

	// Worker pool: fetch -> calculate -> cache -> persist
	workers.Add(cfg.Concurrency)
	for i := 0; i < cfg.Concurrency; i++ {
		go func(workerID int) {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					errs <- fmt.Errorf("worker %d: context canceled: %w", workerID, ctx.Err())
					return
				case username, ok := <-jobs:
					if !ok {
						return
					}

					data, err := fetcher.FetchUserData(ctx, username)
					if err != nil {
						errs <- fmt.Errorf("worker %d: failed to fetch data for user %s: %w", workerID, username, err)
						continue
					}

					if data == nil {
						errs <- fmt.Errorf("worker %d: received nil user data for %s", workerID, username)
						continue
					}

					stats := CalculateStats(data.Repositories)
					cache.Set(username, &stats, cfg.CacheTTL)

					select {
					case results <- username:
					case <-ctx.Done():
						errs <- fmt.Errorf("worker %d: context canceled while sending result", workerID)
						return
					}
				}
			}
		}(i)
	}

	// Producer: load usernames and dispatch in batches
	producer.Add(1)
	go func() {
		defer producer.Done()
		defer close(jobs)

		usernames, err := repo.GetAllUsernames()
		if err != nil {
			errs <- fmt.Errorf("producer: failed to fetch all usernames: %w", err)
			return
		}

		if len(usernames) == 0 {
			return
		}

		for i := 0; i < len(usernames); i += cfg.BatchSize {
			// Check context cancellation before processing batch
			select {
			case <-ctx.Done():
				errs <- fmt.Errorf("producer: context canceled during batch processing: %w", ctx.Err())
				return
			default:
			}

			end := i + cfg.BatchSize
			if end > len(usernames) {
				end = len(usernames)
			}
			batch := usernames[i:end]

			// Send batch jobs
			for _, u := range batch {
				select {
				case <-ctx.Done():
					errs <- fmt.Errorf("producer: context canceled while sending jobs: %w", ctx.Err())
					return
				case jobs <- u:
				}
			}
		}
	}()

	// Wait for producer and workers to complete, then close channels
	go func() {
		producer.Wait()
		workers.Wait()
		close(results)
		close(errs)
	}()
	return results, errs
}
