package kafka

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/domain"
)

func FromDomainBannerInfoToPayload(bf domain.LTBannerInfo) Payload {
	return Payload{
		Username:    bf.Username,
		BannerType:  domain.BannerTypesBackward[bf.BannerType],
		StoragePath: bf.UrlPath,
		Stats:       FromDomainUserStats(bf.Stats),
		FetchedAt:   bf.Stats.FetchedAt,
	}
}

func FromDomainUserStats(us domain.GithubUserStats) Stats {
	return Stats{
		TotalRepos:    us.TotalRepos,
		OriginalRepos: us.OriginalRepos,
		ForkedRepos:   us.ForkedRepos,
		TotalStars:    us.TotalStars,
		TotalForks:    us.TotalForks,
		Languages:     us.Languages,
	}
}

type Stats struct {
	TotalRepos    int            `json:"total_repos"`
	OriginalRepos int            `json:"original_repos"`
	ForkedRepos   int            `json:"forked_repos"`
	TotalStars    int            `json:"total_stars"`
	TotalForks    int            `json:"total_forks"`
	Languages     map[string]int `json:"languages"`
}

type Payload struct {
	Username    string    `json:"username"`
	BannerType  string    `json:"banner_type"`
	StoragePath string    `json:"storage_path"`
	Stats       Stats     `json:"stats"`
	FetchedAt   time.Time `json:"fetched_at"`
}

type GithubBannerInfoEvent struct {
	EventType    string    `json:"event_type"`
	EventVersion int       `json:"event_version"`
	ProducedAt   time.Time `json:"produced_at"`
	Payload      Payload   `json:"payload"`
}
