package banners

import (
	"github.com/hurtki/github-banners/api/internal/domain"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
)

func (r *PostgresRepo) bannerTypeToDB(bt domain.BannerType) (string, error) {
	fn := "internal.repo.banners.PostgresRepo.bannerTypeToDB"
	v, ok := domain.BannerTypesBackward[bt]
	if !ok {
		if r.logger != nil {
			r.logger.Error("unknown banner type", "source", fn, "banner_type", bt)
		}
		return "", repoerr.ErrRepoInternal{Note: "unknown banner type"}
	}
	return v, nil
}

func (r *PostgresRepo) bannerTypeFromDB(v string) (domain.BannerType, error) {
	fn := "internal.repo.banners.PostgresRepo.bannerTypeFromDB"
	bt, ok := domain.BannerTypes[v]
	if !ok {
		if r.logger != nil {
			r.logger.Error("unknown banner type", "source", fn, "banner_type", v)
		}
		return 0, repoerr.ErrRepoInternal{Note: "unknown banner type"}
	}
	return bt, nil
}
