package banners

import (
	"github.com/hurtki/github-banners/api/internal/domain"
	repoerr "github.com/hurtki/github-banners/api/internal/repo"
)

func bannerTypeToDB(bt domain.BannerType) (string, error) {
	v, ok := domain.BannerTypesBackward[bt]
	if !ok {
		return "", repoerr.ErrRepoInternal{Note: "unknown banner type"}
	}
	return v, nil
}

func bannerTypeFromDB(v string) (domain.BannerType, error) {
	bt, ok := domain.BannerTypes[v]
	if !ok {
		return 0, repoerr.ErrRepoInternal{Note: "unknown banner type"}
	}
	return bt, nil
}
