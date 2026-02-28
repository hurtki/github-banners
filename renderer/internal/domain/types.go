package domain

import (
	"time"
)

type BannerType string

const (
	BannerTypeDefault BannerType = "default"
	BannerTypeDark    BannerType = "dark"
)

type BannerInfo struct {
	Username   string
	BannerType BannerType
	URLPath    string
	Stats      GithubUserStats
}

type GithubUserStats struct {
	TotalRepos    int
	OriginalRepos int
	ForkedRepos   int
	TotalStars    int
	TotalForks    int
	Languages     map[string]int
	FetchedAt     time.Time
}
