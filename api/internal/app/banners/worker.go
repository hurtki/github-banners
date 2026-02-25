package banners_worker

import (
	"context"
	"sync"
	"time"

	longterm "github.com/hurtki/github-banners/api/internal/domain/long-term"
	"github.com/hurtki/github-banners/api/internal/logger"
)

type UpdateAllFunc func(context.Context, longterm.UpdateAllConfig) (<-chan longterm.Result, error)

type BannersWorker struct {
	logger    logger.Logger
	updateAll UpdateAllFunc
	interval  time.Duration
	cfg       longterm.UpdateAllConfig

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func NewBannersWorker(logger logger.Logger, updateAllFunc UpdateAllFunc, interval time.Duration, cfg longterm.UpdateAllConfig) *BannersWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &BannersWorker{
		logger:    logger.With("service", "banner-updater-worker"),
		interval:  interval,
		updateAll: updateAllFunc,
		cfg:       cfg,
		ctx:       ctx,
		cancel:    cancel,
		wg:        sync.WaitGroup{},
	}
}

func (w *BannersWorker) Close(ctx context.Context) error {
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

func (w *BannersWorker) Start() {
	w.wg.Go(w.run)
}

func (w *BannersWorker) run() {
	if w == nil || w.updateAll == nil {
		w.logger.Warn("updateAllFunc or worker object is nil, worker didn't start")
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("started", "interval", w.interval.String())

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			ctx, cancel := context.WithTimeout(w.ctx, time.Second*5)
			defer cancel()
			resultsCh, err := w.updateAll(ctx, w.cfg)
			if err != nil {
				if err == context.Canceled {
					return
				}
				w.logger.Error("can't start banners update process", "err", err)
				continue
			}

			success := 0
			errors := 0

			for resultsCh != nil {
				select {
				case <-w.ctx.Done():
					return
				case res, ok := <-resultsCh:
					if !ok {
						resultsCh = nil
						continue
					}
					if res.Err != nil {
						errors++
						w.logger.Error("can't update", "username", res.Meta.Username, "type", res.Meta.BannerType, "url-path", res.Meta.UrlPath, "err", res.Err)
					} else {
						success++
						w.logger.Debug("updated", "username", res.Meta.Username, "type", res.Meta.BannerType, "url-path", res.Meta.UrlPath)
					}
				}
			}

			w.logger.Info("finshed scheduled update", "success", success, "errors", errors, "duration", time.Since(start).String())
			ticker.Reset(w.interval)
		}
	}

}
