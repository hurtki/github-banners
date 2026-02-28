package userstats_worker

import (
	"context"
	"sync"
	"time"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type RefreshAllFunc func(ctx context.Context, cfg userstats.WorkerConfig) (<-chan string, <-chan error)

type StatsWorker struct {
	refreshAll RefreshAllFunc
	interval   time.Duration
	logger     logger.Logger
	cfg        userstats.WorkerConfig

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func NewStatsWorker(refreshAll RefreshAllFunc, interval time.Duration, logger logger.Logger, cfg userstats.WorkerConfig) *StatsWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &StatsWorker{
		refreshAll: refreshAll,
		interval:   interval,
		logger:     logger.With("service", "stats-updater-worker"),
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
	}
}

func (w *StatsWorker) Close(ctx context.Context) error {
	w.cancel()
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		w.logger.Warn("couldn't shutdown in time, exiting", "ctxErr", ctx.Err())
		return ctx.Err()
	case <-done:
		w.logger.Info("successfully shutted down")
		return nil
	}
}

func (w *StatsWorker) Start() {
	w.wg.Go(w.run)
}

func (w *StatsWorker) run() {
	if w == nil || w.refreshAll == nil {
		w.logger.Warn("refreshAll func or worker object is nil; worker ")
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("started", "interval", w.interval.String())

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(w.ctx, time.Second*5)
			defer cancel()

			resultsCh, errorsCh := w.refreshAll(ctx, w.cfg)
			success := 0
			errors := 0

			for resultsCh != nil || errorsCh != nil {
				select {
				case <-w.ctx.Done():
					return

				case username, ok := <-resultsCh:
					if !ok {
						resultsCh = nil
						continue
					}
					success++
					w.logger.Debug("refreshed", "username", username)

				case err, ok := <-errorsCh:
					if !ok {
						errorsCh = nil
						continue
					}
					if err != nil {
						errors++
						w.logger.Debug("error when updating", "err", err)
					}
				}
			}
			w.logger.Info("finshed scheduled update", "success", success, "errors", errors)
		}
	}
}
