package banners_worker

import (
	"context"
	"time"

	longterm "github.com/hurtki/github-banners/api/internal/domain/long-term"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type UpdateAllFunc func(context.Context, longterm.UpdateAllConfig) (<-chan longterm.Result, error)

type BannersWorker struct {
	logger        logger.Logger
	updateAllFunc UpdateAllFunc
	interval      time.Duration
	cfg           longterm.UpdateAllConfig
}

func NewBannersWorker(logger logger.Logger, updateAllFunc UpdateAllFunc, interval time.Duration, cfg longterm.UpdateAllConfig) *BannersWorker {
	return &BannersWorker{
		logger:        logger.With("service", "banner-updater-worker"),
		interval:      interval,
		updateAllFunc: updateAllFunc,
		cfg:           cfg,
	}
}

func (w *BannersWorker) Start(ctx context.Context) {
	if w == nil || w.updateAllFunc == nil {
		w.logger.Warn("updateAllFunc or worker object is nil, worker didn't start")
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("started", "interval", w.interval.String())

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopped, context canceled")
			return
		case <-ticker.C:
			resultsCh, err := w.updateAllFunc(ctx, w.cfg)
			if err != nil {
				w.logger.Error("can't start banners update process", "err", err)
				continue
			}
			success := 0
			errors := 0

			for resultsCh != nil {
				select {
				case <-ctx.Done():
					w.logger.Info("exiting by context")
					return
				case res, ok := <-resultsCh:
					if !ok {
						resultsCh = nil
						continue
					}
					if res.Err != nil {

					} else {
						success++
						w.logger.Debug("updated", "username", res.Meta.Username, "type", res.Meta.BannerType, "url-path", res.Meta.UrlPath)
					}
				}
			}
			w.logger.Info("finshed scheduled update", "success", success, "errors", errors)
		}
	}
}
