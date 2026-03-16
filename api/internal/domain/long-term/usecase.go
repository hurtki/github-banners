package longterm

import (
	"context"
	"errors"
	"path"

	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/repo"
)

type LTBannersUsecase struct {
	bannerRepo             BannerRepo
	updateRequestPublisher UpdateRequestPublisher
	previewService         PreviewService
	storageClient          StorageClient
	statsService           StatsService
}

func NewLTBannersUsecase(
	bannerRepo BannerRepo,
	updateRequestPublisher UpdateRequestPublisher,
	previewService PreviewService,
	storageClient StorageClient,
	statsService StatsService,
) *LTBannersUsecase {
	return &LTBannersUsecase{
		bannerRepo:             bannerRepo,
		updateRequestPublisher: updateRequestPublisher,
		previewService:         previewService,
		storageClient:          storageClient,
		statsService:           statsService,
	}
}

func (u *LTBannersUsecase) CreateBanner(ctx context.Context, in CreateBannerIn) (CreateBannerOut, error) {
	bt, ok := domain.BannerTypes[in.BannerType]
	if !ok {
		return CreateBannerOut{}, ErrInvalidBannerType
	}

	bnrMeta, err := u.bannerRepo.GetBanner(ctx, in.Username, bt)
	if err != nil {
		var errRepoInternal *repo.ErrRepoInternal
		switch {
		case errors.Is(err, repo.ErrNothingFound):
			bnrMeta.Username = in.Username
			bnrMeta.BannerType = bt
			bnrMeta.UrlPath = generateUrlPath(bnrMeta.Username, bnrMeta.BannerType)
			bnrMeta.Active = true
		case errors.As(err, &errRepoInternal):
			// if db internal error occurred, we won't go to next services
			// because, then we could get same thing when saving a new banner and all the work will be useless
			return CreateBannerOut{}, ErrCantCreateBanner
		default:
			return CreateBannerOut{}, ErrCantCreateBanner
		}
	} else {
		if bnrMeta.Active {
			return CreateBannerOut{BannerUrlPath: path.Join("/banners/", bnrMeta.UrlPath)}, nil
		} else {
			bnrMeta.Active = true
		}
	}

	stats, err := u.statsService.GetStats(ctx, in.Username)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return CreateBannerOut{}, ErrUserDoesntExist
		default:
			return CreateBannerOut{}, ErrCantCreateBanner
		}
	}

	// render banner
	bnrInfo := domain.BannerInfo{Username: in.Username, BannerType: bt, Stats: stats}
	bnr, err := u.previewService.GetPreview(ctx, bnrInfo)
	if err != nil {
		return CreateBannerOut{}, ErrCantCreateBanner
	}

	// save rendered banner to storage, so it will be available instantly on returned link
	bannerUrl, err := u.storageClient.SaveBanner(ctx, bnrMeta.UrlPath, string(bnr.Banner))

	if err != nil {
		return CreateBannerOut{}, ErrCantCreateBanner
	}

	// only after all steps, saving to our repo
	err = u.bannerRepo.SaveBanner(ctx, bnrMeta)
	if err != nil {
		return CreateBannerOut{}, ErrCantCreateBanner
	}

	return CreateBannerOut{BannerUrlPath: bannerUrl}, nil
}
