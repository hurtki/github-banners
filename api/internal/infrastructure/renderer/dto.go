package renderer

import "github.com/hurtki/github-banners/api/internal/domain"

// GithubUserBannerInfo is a struct, that describes data that is used to render banner
type GithubUserBannerInfo struct {
	Username   string
	BannerType string
	Stats      domain.GithubUserStats
}

func FromDomainBannerInfo(bi domain.GithubUserBannerInfo) GithubUserBannerInfo {
	return GithubUserBannerInfo{
		Username:   bi.Username,
		BannerType: domain.BannerTypesBackward[bi.BannerType],
		Stats:      bi.Stats,
	}
}

func (i GithubUserBannerInfo) ToBannerPreviewRequest() bannerPreviewRequest {
	return bannerPreviewRequest{
		Username:      i.Username,
		BannerType:    i.BannerType,
		TotalRepos:    i.Stats.TotalRepos,
		OriginalRepos: i.Stats.OriginalRepos,
		ForkedRepos:   i.Stats.ForkedRepos,
		TotalStars:    i.Stats.TotalStars,
		TotalForks:    i.Stats.TotalForks,
		Languages:     i.Stats.Languages,
	}
}

// GithubBanner is rendered banner
type GithubBanner struct {
	Username   string
	BannerType string
	Banner     *[]byte
}

type bannerPreviewRequest struct {
	Username      string         `json:"username"`
	BannerType    string         `json:"banner_type"`
	TotalRepos    int            `json:"total_repos"`
	OriginalRepos int            `json:"original_repos"`
	ForkedRepos   int            `json:"forked_repos"`
	TotalStars    int            `json:"total_stars"`
	TotalForks    int            `json:"total_forks"`
	Languages     map[string]int `json:"languages"`
}
