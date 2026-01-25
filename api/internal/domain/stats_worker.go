package domain

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	SoftTTL = 10 * time.Minute 
	HardTTL = 24 * time.Hour
)

type GithubStatsRepository interface {
	GetUserData(username string) (GithubUserData, error)
	UpdateUserData(userData GithubUserData) error
	GetAllUsernames() ([]string, error)
}

type Cache interface {
	Get(username string) (*CachedStats, bool)
	Set(username string, entry *CachedStats, ttl time.Duration)
	Delete(username string)
}

type UserDataFetcher interface {
	FetchUserData(ctx context.Context, username string) (*GithubUserData, error)
}

type CacheWriter interface {
	Set(username string, stats *GithubUserStats, ttl time.Duration)
}

type UserStatsService struct {
	repo    GithubStatsRepository	
	fetcher	UserDataFetcher
	cache	Cache	
	cfg		WorkerConfig
}

type CachedStats struct {
	Stats		GithubUserStats
	UpdatedAt	time.Time
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

type UserStatsWorker struct {
	service *UserStatsService
	cfg		WorkerConfig
}

func NewUserStatsService( repo GithubStatsRepository, fetcher UserDataFetcher, cache Cache) *UserStatsService {
	return &UserStatsService{
		repo: repo, 
		fetcher: fetcher, 
		cache: cache,
	}
}

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

//fetch api -> save db -> calc stats -> write cache
func (s *UserStatsService) RecalculateAndSync(ctx context.Context, username string) (GithubUserStats, error) {
	//fetching raw data from github
	data, err := s.fetcher.FetchUserData(ctx, username)
	if err != nil {
		return GithubUserStats{}, err
	}

	stats := CalculateStats(data.Repositories)
	//updating database with the new raw data
	if err := s.repo.UpdateUserData(*data); err != nil {
		log.Printf("DB Update failed for %s: %w", username, err)
	}
	s.cache.Set(username, &CachedStats{
		Stats: stats,
		UpdatedAt: time.Now(),
	}, HardTTL)

	return stats, nil
}

func (s *UserStatsService) GetStats(ctx context.Context, username string) (GithubUserStats, error) {
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
	dbData, err := s.repo.GetUserData(username)
	if err == nil {
		stats := CalculateStats(dbData.Repositories)
		s.cache.Set(username, &CachedStats{
			Stats: stats,
			UpdatedAt: time.Now(),
		}, HardTTL)
		return stats, nil
	}

	return s.RecalculateAndSync(ctx, username)
}

func (w *UserStatsService) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	log.Printf("Worker started running every %w", interval)

	w.RunSingleCycle(ctx)

	for {
		select {
			case <-ctx.Done():
			log.Println("Worker stopped: context canceled")
		}
	}
}

func (w *UserStatsService) RunSingleCycle(ctx context.Context) {
	log.Println("worker: starting sync cycle...")
	
	usernames, err := w.repo.GetAllUsernames()
	if err != nil || len(usernames) == 0 {
		return
	}

	jobs := make(chan string, w.cfg.Concurrency)
	var wg sync.WaitGroup

	for i:= 0; i < w.cfg.Concurrency; i++ {
		wg.Add(1)
		go func ()  {
			defer wg.Done()
			for username := range jobs {
				w.RecalculateAndSync(ctx, username)
			}
		}()
	}

	for _, u := range usernames {
		jobs <- u
	}
	close(jobs)
	wg.Wait()
}

func RunUserStatsWorker(ctx context.Context, service *UserStatsService, cfg WorkerConfig) (<-chan string, <-chan error) {
	if cfg.BatchSize <= 0 {cfg.BatchSize = 10}
	if cfg.Concurrency <= 0 {cfg.Concurrency = 4}

	jobs := make(chan string, cfg.Concurrency)
	results := make(chan string, cfg.BatchSize)
	errs := make(chan error, cfg.Concurrency*2)

	var workers sync.WaitGroup
	var producer sync.WaitGroup

	workers.Add(cfg.Concurrency)
	for i := 0; i< cfg.Concurrency; i++ {
		go func() {
			defer workers.Done()
			for{
				select {
				case <-ctx.Done():
					return
				case username, ok := <-jobs:
					if !ok { return}
					
					_, err := service.RecalculateAndSync(ctx, username)
					if err != nil {
						errs <- fmt.Errorf("worker: failed to update for %s: %w", username, err)
						continue
					}

					results <- username
				}
			}
		}()
	}

	producer.Add(1)
	go func() {
		defer producer.Done()
		defer close(jobs)

		usernames, err := service.repo.GetAllUsernames()
		if err != nil {
			errs <- fmt.Errorf("producer: %w", err)
			return
		}
		for _, u := range usernames {
			select {
			case <-ctx.Done(): return
			case jobs <- u:
			}
		}
	}()
	
	//cleanup
	go func() {
		producer.Wait()
		workers.Wait()
		close(results)
		close(errs)
	}()

	return nil, nil

}

//deleting the flow
func(s *UserStatsService) PurgeUser(username string){
	s.cache.Delete(username)
}
