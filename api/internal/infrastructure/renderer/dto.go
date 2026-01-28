package renderer

import "github.com/hurtki/github-banners/api/internal/domain"

// GithubUserBannerInfo is a struct, that describes data that is used to render banner
type GithubUserBannerInfo struct {
	Username   string
	BannerType string
	domain.GithubUserStats
}

func (i GithubUserBannerInfo) ToBannerPreviewRequest() bannerPreviewRequest {
	return bannerPreviewRequest{
		Username:      i.Username,
		BannerType:    i.BannerType,
		TotalRepos:    i.TotalRepos,
		OriginalRepos: i.OriginalRepos,
		ForkedRepos:   i.ForkedRepos,
		TotalStars:    i.TotalStars,
		TotalForks:    i.TotalForks,
		Languages:     i.Languages,
	}
}

// GithubBanner is rendered banner
type GithubBanner struct {
	Username   string
	BannerType string
	Banner     []byte
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
