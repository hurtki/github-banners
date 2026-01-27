package userstats

import (
	"context"
	"time"

	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type StatsWorker struct {
	refreshAll RefreshAllFunc
	interval   time.Duration
	logger     logger.Logger
	cfg        userstats.WorkerConfig
}

type RefreshAllFunc func(ctx context.Context, cfg userstats.WorkerConfig) (<-chan string, <-chan error)

func NewStatsWorker(refreshAll RefreshAllFunc, interval time.Duration, logger logger.Logger, cfg userstats.WorkerConfig) *StatsWorker {
	return &StatsWorker{
		refreshAll: refreshAll,
		interval:   interval,
		logger:     logger.With("service", "stats-updater-worker"),
		cfg:        cfg,
	}
}

func (w *StatsWorker) Start(ctx context.Context) {
	if w == nil || w.refreshAll == nil {
		w.logger.Warn("stats worker: refreshAll is nil; worker not started")
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("stats worker: started", "interval", w.interval.String())

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stats worker: stopped, context canceled")
			return
		case <-ticker.C:
			resultsCh, errorsCh := w.refreshAll(ctx, w.cfg)
			success := 0
			errors := 0

			for resultsCh != nil || errorsCh != nil {
				select {
				case <-ctx.Done():
					w.logger.Info("exiting stats worker by context")
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
