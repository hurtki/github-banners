package userstats

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type WorkerConfig struct {
	BatchSize   int
	Concurrency int
}

type UserStatsRefresher interface {
	RecalculateAndSync(context.Context, string) (interface{}, error)
	GetAllUsernames() ([]string, error)
}

func RunUserStatsWorker(ctx context.Context, service UserStatsRefresher, cfg WorkerConfig) (<-chan string, <-chan error) {
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

		usernames, err := service.GetAllUsernames()
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
	}()

	//cleanup
	go func() {
		producer.Wait()
		workers.Wait()
		close(results)
		close(errs)
	}()

	return results, errs

}

type StatsWorker struct {
	refreshAll func(context.Context) (<-chan string, <-chan error)
	interval   time.Duration
}

func (w *StatsWorker) Start(ctx context.Context) {
	if w == nil || w.refreshAll == nil {
		log.Println("stats worker: refreshAll is nil; worker not started")
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("stats worker: started, interval=%s", w.interval.String())

	for {
		select {
		case <-ctx.Done():
			log.Println("stats worker: stopped, context canceled")
			return
		case <-ticker.C:
			resultsCh, errorsCh := w.refreshAll(ctx)

			// Drain results and errors in background so Start
			// keeps responding to context cancellation.
			go func() {
				for username := range resultsCh {
					log.Printf("stats worker: refreshed %s", username)
				}
			}()

			go func() {
				for err := range errorsCh {
					if err != nil {
						log.Printf("stats worker: error: %v", err)
					}
				}
			}()
		}
	}
}
