package longterm

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hurtki/github-banners/api/internal/domain"
)

type UpdateAllConfig struct {
	Concurrency int
}

type Result struct {
	Meta domain.LTBannerMetadata
	Err  error
}

func (u *LTBannersUsecase) UpdateAll(ctx context.Context, cfg UpdateAllConfig) (<-chan Result, error) {
	jobs, err := u.bannerRepo.GetActiveBanners(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get active banners from repo: %w", err)
	}

	resultsCh := make(chan Result)
	workersWg := sync.WaitGroup{}
	jobsCh := make(chan domain.LTBannerMetadata)

	workersWg.Add(cfg.Concurrency)
	for range cfg.Concurrency {
		go func() {
			defer workersWg.Done()
			for job := range jobsCh {
				err := u.updateOne(ctx, job)
				if err != nil {
					resultsCh <- Result{job, fmt.Errorf("can't update banner: %w", err)}
					continue
				}
				resultsCh <- Result{job, nil}
			}
		}()
	}

	go func() {
		for _, jb := range jobs {
			select {
			case <-ctx.Done():
				close(jobsCh)
				return
			case jobsCh <- jb:
			}
		}
		close(jobsCh)
	}()

	go func() {
		workersWg.Wait()
		close(resultsCh)
	}()

	return resultsCh, nil
}

func (u *LTBannersUsecase) updateOne(ctx context.Context, bannerMeta domain.LTBannerMetadata) error {
	stats, err := u.statsService.GetStats(ctx, bannerMeta.Username)
	if err != nil {
		// if user is not on github -> deactivate his banner
		if errors.Is(err, domain.ErrNotFound) {
			u.bannerRepo.DeactivateBanner(ctx, bannerMeta.Username, bannerMeta.BannerType)
			return fmt.Errorf("user not found on github, deactivating banner: %w", err)
		}
		return fmt.Errorf("can't get user's github stats: %w", err)
	}

	ltBannerInfo := domain.LTBannerInfo{
		BannerInfo: domain.BannerInfo{
			Username:   bannerMeta.Username,
			BannerType: bannerMeta.BannerType,
			Stats:      stats,
		},
		UrlPath: bannerMeta.UrlPath,
	}
	err = u.updateRequestPublisher.Publish(ctx, ltBannerInfo)
	if err != nil {
		return fmt.Errorf("can't publish update request: %w", err)
	}
	return nil
}
