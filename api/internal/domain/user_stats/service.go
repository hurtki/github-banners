package userstats

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

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
	dbData, err := s.repo.GetUserData(username)
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
	if err := s.repo.UpdateUserData(*data); err != nil {
		log.Printf("DB Update failed for %s: %w", username, err)
	}
	s.cache.Set(username, &CachedStats{
		Stats:     stats,
		UpdatedAt: time.Now(),
	}, HardTTL)

	return stats, nil
}

func (w *UserStatsService) RefreshAll(ctx context.Context) {
	log.Println("worker: starting sync cycle...")

	usernames, err := w.repo.GetAllUsernames()
	if err != nil || len(usernames) == 0 {
		return
	}

	jobs := make(chan string, w.cfg.Concurrency)
	var wg sync.WaitGroup

	for i := 0; i < w.cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
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
