package domain

import (
	"time"
)

type GithubRepository struct {
	ID            int64
	OwnerUsername string
	PushedAt      *time.Time
	UpdatedAt     *time.Time
	Language      *string
	StarsCount    int
	Fork          bool
	ForksCount    int
}

type GithubUserData struct {
	Username     string
	Name         *string
	Company      *string
	Location     *string
	Bio          *string
	PublicRepos  int
	Followers    int
	Following    int
	Repositories []GithubRepository
	FetchedAt    time.Time
}

type GithubUserStats struct {
	TotalRepos    int
	OriginalRepos int
	ForkedRepos   int
	TotalStars    int
	TotalForks    int
	Languages     map[string]int
}

type ServiceConfig struct {
	CacheTTL       time.Duration
	RequestTimeout time.Duration
}

type BannerType int

const (
	TypeWide = iota
)

var BannerTypes map[string]BannerType = map[string]BannerType{"wide": TypeWide}
var BannerTypesBackward map[BannerType]string = map[BannerType]string{TypeWide: "wide"}

type GithubBanner struct {
	Username   string
	BannerType BannerType
	Banner     *[]byte
}

type GithubUserBannerInfo struct {
	Username   string
	BannerType BannerType
	Stats      GithubUserStats
}
