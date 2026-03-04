package renderer

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

// GithubUserBannerInfo is a struct, that describes data that is used to render banner
type GithubUserBannerInfo struct {
	Username   string
	BannerType string
	Stats      domain.GithubUserStats
}

func FromDomainBannerInfo(bi domain.BannerInfo) GithubUserBannerInfo {
	return GithubUserBannerInfo{
		Username:   bi.Username,
		BannerType: domain.BannerTypesBackward[bi.BannerType],
		Stats:      bi.Stats,
	}
}

func (i GithubUserBannerInfo) ToBannerPreviewRequest() bannerPreviewRequest {
	return bannerPreviewRequest{
		Username:   i.Username,
		BannerType: i.BannerType,
		Stats: bannerPreviewStats{
			TotalRepos:    i.Stats.TotalRepos,
			OriginalRepos: i.Stats.OriginalRepos,
			ForkedRepos:   i.Stats.ForkedRepos,
			TotalStars:    i.Stats.TotalStars,
			TotalForks:    i.Stats.TotalForks,
			Languages:     i.Stats.Languages,
		},
		FetchedAt: i.Stats.FetchedAt,
	}
}

type bannerPreviewRequest struct {
	Username   string             `json:"username"`
	BannerType string             `json:"banner_type"`
	Stats      bannerPreviewStats `json:"stats"`
	FetchedAt  time.Time          `json:"fetched_at"`
}

type bannerPreviewStats struct {
	TotalRepos    int            `json:"totalRepos"`
	OriginalRepos int            `json:"originalRepos"`
	ForkedRepos   int            `json:"forkedRepos"`
	TotalStars    int            `json:"totalStars"`
	TotalForks    int            `json:"totalForks"`
	Languages     map[string]int `json:"languages"`
}
