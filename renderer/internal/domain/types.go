package domain

import (
	"context"
	"time"
)

type BannerInfo struct {
    Username    string
    BannerType  string
    Profile     GithubUserData
    Stats       GithubUserStats
}

type RenderedBanner struct {
    Filename    string
    Data        []byte
}

type GithubUserData struct {
    Username    string
    Name        *string
    Company     *string
    Location    *string
    Bio         *string
    PublicRepos int
    Followers   int
    Following   int
    Repositories []GithubRepository
    FetchedAt   time.Time
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

type GithubRepository struct {
    ID            int64
    OwnerUsername string
    PushedAt      *time.Time
    UpdatedAt     *time.Time
    Language      *string
    StarsCount    int
    Fork          int
    ForksCount    int
}

type BannerStorage interface {
    Save(ctx context.Context, banner RenderedBanner) error
}
