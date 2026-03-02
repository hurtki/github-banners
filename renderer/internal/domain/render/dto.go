package render

import "github.com/hurtki/github-banners/renderer/internal/domain"

type UpdateBannerIn struct {
	Username   string
	BannerType string
	URLPath    string
	Stats      domain.GithubUserStats
}
